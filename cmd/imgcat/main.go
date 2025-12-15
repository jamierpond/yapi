package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/sys/unix"
	"golang.org/x/term"

	// Register decoders
	_ "image/gif"
)

// --- Enums and Types ---

type Unit int

const (
	UnitCells Unit = iota
	UnitPixels
	UnitPercent
	UnitAuto
)

type ITermDimension struct {
	Value float64
	Unit  Unit
}

func (d *ITermDimension) String() string {
	switch d.Unit {
	case UnitPixels:
		return fmt.Sprintf("%gpx", d.Value)
	case UnitPercent:
		return fmt.Sprintf("%g%%", d.Value)
	case UnitCells:
		return fmt.Sprintf("%g", d.Value)
	default:
		return "auto"
	}
}

func (d *ITermDimension) Set(s string) error {
	if s == "auto" {
		d.Unit = UnitAuto
		return nil
	}
	if strings.HasSuffix(s, "px") {
		d.Unit = UnitPixels
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		if err != nil {
			return err
		}
		d.Value = v
		return nil
	}
	if strings.HasSuffix(s, "%") {
		d.Unit = UnitPercent
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return err
		}
		d.Value = v
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	d.Value = v
	d.Unit = UnitCells
	return nil
}

// Helper to convert dimension to pixels based on context
func (d ITermDimension) ToPixels(cellSize, limit int) float64 {
	switch d.Unit {
	case UnitPixels:
		return d.Value
	case UnitPercent:
		return (d.Value / 100.0) * float64(limit*cellSize)
	case UnitCells:
		return d.Value * float64(cellSize)
	default:
		return 0 // Auto
	}
}

type ImagePosition struct {
	X, Y int
}

func (p *ImagePosition) String() string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}

func (p *ImagePosition) Set(s string) error {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return errors.New("expected x,y")
	}
	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid x: %w", err)
	}
	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid y: %w", err)
	}
	p.X = x
	p.Y = y
	return nil
}

type ImageDimension struct {
	Width, Height int
}

func (d *ImageDimension) String() string {
	return fmt.Sprintf("%dx%d", d.Width, d.Height)
}

func (d *ImageDimension) Set(s string) error {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return errors.New("expected WxH")
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid width: %w", err)
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid height: %w", err)
	}
	d.Width = w
	d.Height = h
	return nil
}

type TmuxPassthru string

const (
	TmuxDisable TmuxPassthru = "disable"
	TmuxEnable  TmuxPassthru = "enable"
	TmuxDetect  TmuxPassthru = "detect"
)

func (t *TmuxPassthru) String() string { return string(*t) }
func (t *TmuxPassthru) Set(s string) error {
	val := TmuxPassthru(strings.ToLower(s))
	switch val {
	case TmuxDisable, TmuxEnable, TmuxDetect:
		*t = val
		return nil
	}
	return errors.New("must be disable, enable, or detect")
}

func (t TmuxPassthru) Enabled() bool {
	switch t {
	case TmuxEnable:
		return true
	case TmuxDisable:
		return false
	case TmuxDetect:
		return os.Getenv("TMUX") != ""
	default:
		return false
	}
}

func (t TmuxPassthru) Encode(content string) string {
	if t.Enabled() {
		var sb strings.Builder
		sb.WriteString("\x1bPtmux;")
		for _, c := range content {
			if c == '\x1b' {
				sb.WriteString("\x1b\x1b")
			} else {
				sb.WriteRune(c)
			}
		}
		sb.WriteString("\x1b\\")
		return sb.String()
	}
	return content
}

type ResampleFormat string

const (
	FormatPng   ResampleFormat = "png"
	FormatJpeg  ResampleFormat = "jpeg"
	FormatInput ResampleFormat = "input"
)

func (f *ResampleFormat) String() string { return string(*f) }
func (f *ResampleFormat) Set(s string) error {
	val := ResampleFormat(strings.ToLower(s))
	switch val {
	case FormatPng, FormatJpeg, FormatInput:
		*f = val
		return nil
	}
	return errors.New("must be png, jpeg, or input")
}

type ResampleFilter string

const (
	FilterNearest    ResampleFilter = "nearest"
	FilterTriangle   ResampleFilter = "triangle"
	FilterCatmullRom ResampleFilter = "catmull-rom"
	FilterGaussian   ResampleFilter = "gaussian"
	FilterLanczos3   ResampleFilter = "lanczos3"
)

func (f *ResampleFilter) String() string { return string(*f) }
func (f *ResampleFilter) Set(s string) error {
	val := ResampleFilter(strings.ToLower(s))
	switch val {
	case FilterNearest, FilterTriangle, FilterCatmullRom, FilterGaussian, FilterLanczos3:
		*f = val
		return nil
	}
	return errors.New("unknown filter type")
}

// --- Configuration Struct ---

type Config struct {
	Width                 ITermDimension
	Height                ITermDimension
	NoPreserveAspectRatio bool
	Position              *ImagePosition
	NoMoveCursor          bool
	Hold                  bool
	TmuxPassthru          TmuxPassthru
	MaxPixels             int
	NoResample            bool
	ResampleFormat        ResampleFormat
	ResampleFilter        ResampleFilter
	Resize                *ImageDimension
	ShowResampleTiming    bool
	FileName              string
}

// --- Logic ---

type ImageInfo struct {
	Width  int
	Height int
	Format string // "png", "jpeg", "gif", etc.
}

func getImageDimensions(data []byte) (ImageInfo, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ImageInfo{}, err
	}
	return ImageInfo{Width: cfg.Width, Height: cfg.Height, Format: format}, nil
}

// Maps the CLI filter enum to the draw package scaler
func (c *Config) getScaler() draw.Interpolator {
	switch c.ResampleFilter {
	case FilterNearest:
		return draw.NearestNeighbor
	case FilterTriangle:
		return draw.BiLinear // Closest approximation
	case FilterCatmullRom:
		return draw.CatmullRom
	case FilterGaussian:
		return draw.ApproxBiLinear // Approximation
	case FilterLanczos3:
		// Go's x/image/draw doesn't export Lanczos3 directly as a named var often used,
		// but CatmullRom is high quality cubic. We'll fallback to CatmullRom or define custom kernel if strictly needed.
		// For simplicity/perf balance in Go, CatmullRom is usually the standard "high quality".
		return draw.CatmullRom
	default:
		return draw.CatmullRom
	}
}

func (c *Config) resizeImage(data []byte, targetW, targetH int, info ImageInfo) ([]byte, ImageInfo, error) {
	start := time.Now()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, ImageInfo{}, fmt.Errorf("decoding image: %w", err)
	}
	if c.ShowResampleTiming {
		fmt.Fprintf(os.Stderr, "loading image took %v for %d bytes\n", time.Since(start), len(data))
	}

	start = time.Now()
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	scaler := c.getScaler()
	scaler.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	if c.ShowResampleTiming {
		fmt.Fprintf(os.Stderr, "resizing took %v\n", time.Since(start))
	}

	start = time.Now()
	var outBuf bytes.Buffer
	outFormat := c.ResampleFormat
	if outFormat == FormatInput {
		// Map generic input format names to what we can encode
		switch info.Format {
		case "jpeg", "jpg":
			outFormat = FormatJpeg
		case "png":
			outFormat = FormatPng
		default:
			// Fallback if input was gif/webp etc and we are forced to re-encode
			outFormat = FormatPng
		}
	}

	var newFormatStr string
	switch outFormat {
	case FormatJpeg:
		err = jpeg.Encode(&outBuf, dst, nil)
		newFormatStr = "jpeg"
	default:
		err = png.Encode(&outBuf, dst)
		newFormatStr = "png"
	}

	if err != nil {
		return nil, ImageInfo{}, fmt.Errorf("encoding resampled image: %w", err)
	}

	newInfo := ImageInfo{
		Width:  targetW,
		Height: targetH,
		Format: newFormatStr,
	}

	if c.ShowResampleTiming {
		fmt.Fprintf(os.Stderr, "encoding took %v to produce %d bytes -> %+v\n", time.Since(start), outBuf.Len(), newInfo)
	}

	return outBuf.Bytes(), newInfo, nil
}

func (c *Config) getImageData() ([]byte, ImageInfo, error) {
	var data []byte
	var err error

	if c.FileName != "" {
		data, err = os.ReadFile(c.FileName)
		if err != nil {
			return nil, ImageInfo{}, fmt.Errorf("reading file %s: %w", c.FileName, err)
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, ImageInfo{}, fmt.Errorf("reading stdin: %w", err)
		}
	}

	info, err := getImageDimensions(data)
	if err != nil {
		return nil, ImageInfo{}, err
	}

	// 1. Explicit Resize
	if c.Resize != nil {
		data, info, err = c.resizeImage(data, c.Resize.Width, c.Resize.Height, info)
		if err != nil {
			return nil, ImageInfo{}, err
		}
	}

	// 2. Max Pixels limit
	totalPixels := info.Width * info.Height
	if !c.NoResample && totalPixels > c.MaxPixels {
		scale := float64(totalPixels) / float64(c.MaxPixels)
		// scale is area factor, dimension factor is sqrt(scale)
		dimScale := math.Sqrt(scale)

		targetW := int(float64(info.Width) / dimScale)
		targetH := int(float64(info.Height) / dimScale)
		// Ensure at least 1px
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		data, info, err = c.resizeImage(data, targetW, targetH, info)
		if err != nil {
			return nil, ImageInfo{}, err
		}
	}

	return data, info, nil
}

// Terminal Size handling
type ScreenSize struct {
	Cols   int
	Rows   int
	XPixel int
	YPixel int
}

func getScreenSize() (ScreenSize, error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return ScreenSize{}, err
	}
	return ScreenSize{
		Cols:   int(ws.Col),
		Rows:   int(ws.Row),
		XPixel: int(ws.Xpixel), // Note: might be 0 on some terms
		YPixel: int(ws.Ypixel),
	}, nil
}

func (c *Config) computeImageCellDimensions(info ImageInfo, termSize ScreenSize) (int, int) {
	cellW := termSize.XPixel / termSize.Cols
	cellH := termSize.YPixel / termSize.Rows

	// Fallback to reasonable defaults if ioctl returns 0
	if cellW == 0 {
		cellW = 10
	}
	if cellH == 0 {
		cellH = 20
	}

	pixelWidthLimit := float64(termSize.XPixel)
	pixelHeightLimit := float64(termSize.YPixel)

	// Fallback for limits if pixel dims are missing
	if pixelWidthLimit == 0 {
		pixelWidthLimit = float64(termSize.Cols * cellW)
	}
	if pixelHeightLimit == 0 {
		pixelHeightLimit = float64(termSize.Rows * cellH)
	}

	targetW := c.Width.ToPixels(cellW, termSize.Cols)
	targetH := c.Height.ToPixels(cellH, termSize.Rows)

	aspect := float64(info.Width) / float64(info.Height)

	var finalW, finalH float64

	// Logic to resolve "Auto" vs explicit dimensions
	if c.Width.Unit == UnitAuto && c.Height.Unit == UnitAuto {
		// Native size but ensure fits
		w := float64(info.Width)
		h := float64(info.Height)

		if w > pixelWidthLimit || h > pixelHeightLimit {
			// Fit in box logic (contain)
			xScale := pixelWidthLimit / w
			yScale := pixelHeightLimit / h
			scale := xScale
			if yScale < xScale {
				scale = yScale
			}
			finalW = w * scale
			finalH = h * scale
		} else {
			finalW = w
			finalH = h
		}
	} else if c.Width.Unit != UnitAuto && c.Height.Unit == UnitAuto {
		finalW = targetW
		finalH = targetW / aspect
	} else if c.Width.Unit == UnitAuto && c.Height.Unit != UnitAuto {
		finalH = targetH
		finalW = targetH * aspect
	} else {
		finalW = targetW
		finalH = targetH
		if !c.NoPreserveAspectRatio {
			finalH = finalW / aspect
		}
	}

	// --- FIX START ---
	// Use math.Ceil to ensure we round UP.
	// 4.1 rows of pixels needs 5 rows of text terminal space.
	cols := int(math.Ceil(finalW / float64(cellW)))
	rows := int(math.Ceil(finalH / float64(cellH)))
	// --- FIX END ---

	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return cols, rows
}

func (c *Config) Run() error {
	data, info, err := c.getImageData()
	if err != nil {
		return err
	}

	// Make terminal Raw to query size safely if needed, though we rely on Ioctl mostly.
	// WezTerm Rust code sets Raw then Cooked.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// If stdin isn't a terminal (e.g. piped image), try setting stdout raw
		// This happens if we cat image.png | imgcat
		oldState, err = term.MakeRaw(int(os.Stdout.Fd()))
	}
	if err == nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
		term.Restore(int(os.Stdout.Fd()), oldState)
	}

	termSize, err := getScreenSize()
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	isTmux := os.Getenv("TMUX") != ""
	// Simple ConPTY check
	isConpty := false
	if os.PathSeparator == '\\' {
		// Weak check for windows logic
		isConpty = true
	}

	needsForceCursorMove := !c.NoMoveCursor && c.Position == nil && (isTmux || isConpty) && (termSize.XPixel != 0 && termSize.YPixel != 0)

	// Cursor handling
	saveCursor := "\x1b7"
	restoreCursor := "\x1b8"

	if c.Position != nil {
		// CSI line;col H
		fmt.Printf("%s\x1b[%d;%dH", saveCursor, c.Position.Y+1, c.Position.X+1)
	}

	// Compute dimensions in cells
	_, rows := c.computeImageCellDimensions(info, termSize)

	if needsForceCursorMove {
		// Emit newlines to scroll
		fmt.Print(strings.Repeat("\n", rows))
		// Move back up
		fmt.Printf("\x1b[%dA", rows)
		// Move to col 0
		fmt.Print("\r")
	}

	// Construct OSC 1337
	// File=name=...;size=...;width=...;height=...;preserveAspectRatio=...;inline=1
	var oscBuilder strings.Builder
	oscBuilder.WriteString("\x1b]1337;File=inline=1")
	oscBuilder.WriteString(fmt.Sprintf(";size=%d", len(data)))

	// Rust code always passes width/height params to the protocol even if auto
	// However, standard use of imgcat often lets terminal decide.
	// The Rust code passes: width=self.width, height=self.height.
	if c.Width.Unit != UnitAuto {
		oscBuilder.WriteString(fmt.Sprintf(";width=%s", c.Width.String()))
	}
	if c.Height.Unit != UnitAuto {
		oscBuilder.WriteString(fmt.Sprintf(";height=%s", c.Height.String()))
	}

	if c.NoPreserveAspectRatio {
		oscBuilder.WriteString(";preserveAspectRatio=0")
	} else {
		// Default is usually 1, but wezterm explicitly sends logic
		oscBuilder.WriteString(";preserveAspectRatio=1")
	}

	if c.NoMoveCursor {
		oscBuilder.WriteString(";doNotMoveCursor=1")
	}

	oscBuilder.WriteString(":")
	oscBuilder.WriteString(base64.StdEncoding.EncodeToString(data))
	oscBuilder.WriteString("\a") // Bell terminator

	encoded := c.TmuxPassthru.Encode(oscBuilder.String())
	fmt.Println(encoded)

	if needsForceCursorMove {
		// Move cursor down relative
		fmt.Printf("\x1b[%dB", rows)
	} else if c.Position != nil {
		fmt.Print(restoreCursor)
	}

	if c.Hold {
		// Set Raw Mode and wait for key
		// We need to operate on Stdin. If stdin was the pipe for the image, we might have an issue reading TTY input.
		// Usually imgcat is run as `imgcat file.png` (stdin is tty) or `cat file | imgcat` (stdin is file).
		// If stdin is file, we might need to open /dev/tty to read keys.

		ttyFile, err := os.Open("/dev/tty")
		fd := int(os.Stdin.Fd())
		if err == nil {
			fd = int(ttyFile.Fd())
		}

		state, err := term.MakeRaw(fd)
		if err != nil {
			return fmt.Errorf("failed to enable raw mode for hold: %w", err)
		}
		defer term.Restore(fd, state)

		buf := make([]byte, 1)
		for {
			var n int
			if ttyFile != nil {
				n, _ = ttyFile.Read(buf)
			} else {
				n, _ = os.Stdin.Read(buf)
			}

			if n > 0 {
				b := buf[0]
				// Ctrl+C (3), Ctrl+D (4), Esc (27), Enter (13)
				if b == 3 || b == 4 || b == 27 || b == 13 {
					break
				}
			}
		}
	}

	return nil
}

// --- Main ---

func main() {
	cfg := Config{
		TmuxPassthru:   TmuxDetect,
		MaxPixels:      25000000,
		ResampleFormat: FormatInput,
		ResampleFilter: FilterCatmullRom,
	}

	// Setup Flags
	// We need manual handling to match ValueEnum strings case-insensitively and custom types

	// Helper for Dimensions
	var widthStr, heightStr string
	var posStr string
	var resizeStr string
	var tmuxStr string
	var fmtStr string
	var filterStr string

	flag.StringVar(&widthStr, "width", "auto", "Specify display width (N, Npx, N%)")
	flag.StringVar(&heightStr, "height", "auto", "Specify display height (N, Npx, N%)")
	flag.BoolVar(&cfg.NoPreserveAspectRatio, "no-preserve-aspect-ratio", false, "Do not respect aspect ratio")
	flag.StringVar(&posStr, "position", "", "Set cursor position (x,y)")
	flag.BoolVar(&cfg.NoMoveCursor, "no-move-cursor", false, "Do not move cursor after displaying")
	flag.BoolVar(&cfg.Hold, "hold", false, "Wait for enter/esc/ctrl-c after display")
	flag.StringVar(&tmuxStr, "tmux-passthru", "detect", "Tmux passthrough (enable|disable|detect)")
	flag.IntVar(&cfg.MaxPixels, "max-pixels", 25000000, "Max pixels per frame")
	flag.BoolVar(&cfg.NoResample, "no-resample", false, "Do not resample large images")
	flag.StringVar(&fmtStr, "resample-format", "input", "Format for resampling (png|jpeg|input)")
	flag.StringVar(&filterStr, "resample-filter", "catmull-rom", "Filter (nearest|triangle|catmull-rom|gaussian|lanczos3)")
	flag.StringVar(&resizeStr, "resize", "", "Pre-process resize (WxH)")
	flag.BoolVar(&cfg.ShowResampleTiming, "show-resample-timing", false, "Show timing diagnostics")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Parse Custom Types
	if err := cfg.Width.Set(widthStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --width: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Height.Set(heightStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --height: %v\n", err)
		os.Exit(1)
	}
	if posStr != "" {
		cfg.Position = &ImagePosition{}
		if err := cfg.Position.Set(posStr); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing --position: %v\n", err)
			os.Exit(1)
		}
	}
	if resizeStr != "" {
		cfg.Resize = &ImageDimension{}
		if err := cfg.Resize.Set(resizeStr); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing --resize: %v\n", err)
			os.Exit(1)
		}
	}
	if err := cfg.TmuxPassthru.Set(tmuxStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --tmux-passthru: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.ResampleFormat.Set(fmtStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --resample-format: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.ResampleFilter.Set(filterStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --resample-filter: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) > 0 {
		cfg.FileName = args[0]
	}

	if err := cfg.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

package imgcat

import (
	"bytes"
	"encoding/base64"
	"errors"
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

type Dimension struct {
	Value float64
	Unit  Unit
}

func (d *Dimension) String() string {
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

func (d *Dimension) Set(s string) error {
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

func (d Dimension) ToPixels(cellSize, limit int) float64 {
	switch d.Unit {
	case UnitPixels:
		return d.Value
	case UnitPercent:
		return (d.Value / 100.0) * float64(limit*cellSize)
	case UnitCells:
		return d.Value * float64(cellSize)
	default:
		return 0
	}
}

type Position struct {
	X, Y int
}

func (p *Position) String() string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}

func (p *Position) Set(s string) error {
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

type Size struct {
	Width, Height int
}

func (d *Size) String() string {
	return fmt.Sprintf("%dx%d", d.Width, d.Height)
}

func (d *Size) Set(s string) error {
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

// --- Configuration ---

type Config struct {
	Width                 Dimension
	Height                Dimension
	NoPreserveAspectRatio bool
	Position              *Position
	NoMoveCursor          bool
	Hold                  bool
	TmuxPassthru          TmuxPassthru
	MaxPixels             int
	NoResample            bool
	ResampleFormat        ResampleFormat
	ResampleFilter        ResampleFilter
	Resize                *Size
	ShowResampleTiming    bool
	FileName              string
}

func NewConfig() *Config {
	return &Config{
		TmuxPassthru:   TmuxDetect,
		MaxPixels:      25000000,
		ResampleFormat: FormatInput,
		ResampleFilter: FilterCatmullRom,
	}
}

// --- Internal Types ---

type imageInfo struct {
	Width  int
	Height int
	Format string
}

type screenSize struct {
	Cols   int
	Rows   int
	XPixel int
	YPixel int
}

// --- Logic ---

func getImageDimensions(data []byte) (imageInfo, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return imageInfo{}, err
	}
	return imageInfo{Width: cfg.Width, Height: cfg.Height, Format: format}, nil
}

func (c *Config) getScaler() draw.Interpolator {
	switch c.ResampleFilter {
	case FilterNearest:
		return draw.NearestNeighbor
	case FilterTriangle:
		return draw.BiLinear
	case FilterCatmullRom:
		return draw.CatmullRom
	case FilterGaussian:
		return draw.ApproxBiLinear
	case FilterLanczos3:
		return draw.CatmullRom
	default:
		return draw.CatmullRom
	}
}

func (c *Config) resizeImage(data []byte, targetW, targetH int, info imageInfo) ([]byte, imageInfo, error) {
	start := time.Now()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, imageInfo{}, fmt.Errorf("decoding image: %w", err)
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
		switch info.Format {
		case "jpeg", "jpg":
			outFormat = FormatJpeg
		case "png":
			outFormat = FormatPng
		default:
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
		return nil, imageInfo{}, fmt.Errorf("encoding resampled image: %w", err)
	}

	newInfo := imageInfo{
		Width:  targetW,
		Height: targetH,
		Format: newFormatStr,
	}

	if c.ShowResampleTiming {
		fmt.Fprintf(os.Stderr, "encoding took %v to produce %d bytes -> %+v\n", time.Since(start), outBuf.Len(), newInfo)
	}

	return outBuf.Bytes(), newInfo, nil
}

func (c *Config) getImageData() ([]byte, imageInfo, error) {
	var data []byte
	var err error

	if c.FileName != "" {
		data, err = os.ReadFile(c.FileName)
		if err != nil {
			return nil, imageInfo{}, fmt.Errorf("reading file %s: %w", c.FileName, err)
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, imageInfo{}, fmt.Errorf("reading stdin: %w", err)
		}
	}

	info, err := getImageDimensions(data)
	if err != nil {
		return nil, imageInfo{}, err
	}

	if c.Resize != nil {
		data, info, err = c.resizeImage(data, c.Resize.Width, c.Resize.Height, info)
		if err != nil {
			return nil, imageInfo{}, err
		}
	}

	totalPixels := info.Width * info.Height
	if !c.NoResample && totalPixels > c.MaxPixels {
		scale := float64(totalPixels) / float64(c.MaxPixels)
		dimScale := math.Sqrt(scale)

		targetW := int(float64(info.Width) / dimScale)
		targetH := int(float64(info.Height) / dimScale)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		data, info, err = c.resizeImage(data, targetW, targetH, info)
		if err != nil {
			return nil, imageInfo{}, err
		}
	}

	return data, info, nil
}

func getScreenSize() (screenSize, error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return screenSize{}, err
	}
	return screenSize{
		Cols:   int(ws.Col),
		Rows:   int(ws.Row),
		XPixel: int(ws.Xpixel),
		YPixel: int(ws.Ypixel),
	}, nil
}

func (c *Config) computeImageCellDimensions(info imageInfo, termSize screenSize) (int, int) {
	cellW := termSize.XPixel / termSize.Cols
	cellH := termSize.YPixel / termSize.Rows
	if cellW == 0 {
		cellW = 10
	}
	if cellH == 0 {
		cellH = 20
	}

	pixelWidthLimit := float64(termSize.XPixel)
	pixelHeightLimit := float64(termSize.YPixel)
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

	if c.Width.Unit == UnitAuto && c.Height.Unit == UnitAuto {
		w := float64(info.Width)
		h := float64(info.Height)

		if w > pixelWidthLimit || h > pixelHeightLimit {
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

	cols := int(finalW / float64(cellW))
	rows := int(finalH / float64(cellH))
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
	return c.render(data, info)
}

func (c *Config) render(data []byte, info imageInfo) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
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
	isConpty := os.PathSeparator == '\\'

	needsForceCursorMove := !c.NoMoveCursor && c.Position == nil && (isTmux || isConpty) && (termSize.XPixel != 0 && termSize.YPixel != 0)

	saveCursor := "\x1b7"
	restoreCursor := "\x1b8"

	if c.Position != nil {
		fmt.Printf("%s\x1b[%d;%dH", saveCursor, c.Position.Y+1, c.Position.X+1)
	}

	_, rows := c.computeImageCellDimensions(info, termSize)

	if needsForceCursorMove {
		fmt.Print(strings.Repeat("\n", rows))
		fmt.Printf("\x1b[%dA", rows)
		fmt.Print("\r")
	}

	var oscBuilder strings.Builder
	oscBuilder.WriteString("\x1b]1337;File=inline=1")
	oscBuilder.WriteString(fmt.Sprintf(";size=%d", len(data)))

	if c.Width.Unit != UnitAuto {
		oscBuilder.WriteString(fmt.Sprintf(";width=%s", c.Width.String()))
	}
	if c.Height.Unit != UnitAuto {
		oscBuilder.WriteString(fmt.Sprintf(";height=%s", c.Height.String()))
	}

	if c.NoPreserveAspectRatio {
		oscBuilder.WriteString(";preserveAspectRatio=0")
	} else {
		oscBuilder.WriteString(";preserveAspectRatio=1")
	}

	if c.NoMoveCursor {
		oscBuilder.WriteString(";doNotMoveCursor=1")
	}

	oscBuilder.WriteString(":")
	oscBuilder.WriteString(base64.StdEncoding.EncodeToString(data))
	oscBuilder.WriteString("\a")

	encoded := c.TmuxPassthru.Encode(oscBuilder.String())
	fmt.Println(encoded)

	if needsForceCursorMove {
		fmt.Printf("\x1b[%dB", rows)
	} else if c.Position != nil {
		fmt.Print(restoreCursor)
	}

	if c.Hold {
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
				if b == 3 || b == 4 || b == 27 || b == 13 {
					break
				}
			}
		}
	}

	return nil
}

// PrintConfig is a simplified config for programmatic image printing.
type PrintConfig struct {
	Width  int // Display width in terminal cells (0 = auto)
	Height int // Display height in terminal cells (0 = auto)
}

// Print renders image data to the terminal using iTerm2 inline image protocol.
// This is a convenience function for programmatic use.
func Print(data []byte, cfg PrintConfig) error {
	info, err := getImageDimensions(data)
	if err != nil {
		return err
	}

	c := NewConfig()

	if cfg.Width > 0 {
		c.Width.Value = float64(cfg.Width)
		c.Width.Unit = UnitCells
	}
	if cfg.Height > 0 {
		c.Height.Value = float64(cfg.Height)
		c.Height.Unit = UnitCells
	}

	// Process resampling if needed
	totalPixels := info.Width * info.Height
	if totalPixels > c.MaxPixels {
		scale := float64(totalPixels) / float64(c.MaxPixels)
		dimScale := math.Sqrt(scale)

		targetW := int(float64(info.Width) / dimScale)
		targetH := int(float64(info.Height) / dimScale)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		data, info, err = c.resizeImage(data, targetW, targetH, info)
		if err != nil {
			return err
		}
	}

	return c.render(data, info)
}

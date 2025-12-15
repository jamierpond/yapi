package main

import (
	"flag"
	"fmt"
	"os"

	"yapi.run/cli/internal/imgcat"
)

func main() {
	cfg := imgcat.NewConfig()

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
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file]\n\nDisplay images in the terminal using iTerm2 inline image protocol.\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if err := cfg.Width.Set(widthStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --width: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Height.Set(heightStr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --height: %v\n", err)
		os.Exit(1)
	}
	if posStr != "" {
		cfg.Position = &imgcat.Position{}
		if err := cfg.Position.Set(posStr); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing --position: %v\n", err)
			os.Exit(1)
		}
	}
	if resizeStr != "" {
		cfg.Resize = &imgcat.Size{}
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

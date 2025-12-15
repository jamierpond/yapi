package main

import (
	"flag"
	"fmt"
	"os"

	// Replace with your actual module path
	"yapi.run/cli/internal/imgcat"
)

func main() {
	cfg := imgcat.NewConfig()

	// Temporary string vars for custom Set() parsing
	var widthStr, heightStr string
	var posStr string
	var resizeStr string
	var tmuxStr string
	var fmtStr string
	var filterStr string

	// Define Flags
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
		fmt.Fprintf(os.Stderr, "Usage of %s [options] [file]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Apply parsed strings to Config custom types
	if err := cfg.Width.Set(widthStr); err != nil {
		fail("Error parsing --width: %v", err)
	}
	if err := cfg.Height.Set(heightStr); err != nil {
		fail("Error parsing --height: %v", err)
	}
	if posStr != "" {
		cfg.Position = &imgcat.ImagePosition{}
		if err := cfg.Position.Set(posStr); err != nil {
			fail("Error parsing --position: %v", err)
		}
	}
	if resizeStr != "" {
		cfg.Resize = &imgcat.ImageDimension{}
		if err := cfg.Resize.Set(resizeStr); err != nil {
			fail("Error parsing --resize: %v", err)
		}
	}
	if err := cfg.TmuxPassthru.Set(tmuxStr); err != nil {
		fail("Error parsing --tmux-passthru: %v", err)
	}
	if err := cfg.ResampleFormat.Set(fmtStr); err != nil {
		fail("Error parsing --resample-format: %v", err)
	}
	if err := cfg.ResampleFilter.Set(filterStr); err != nil {
		fail("Error parsing --resample-filter: %v", err)
	}

	// Handle Input File
	args := flag.Args()
	if len(args) > 0 {
		cfg.FileName = args[0]
	}

	// Run
	if err := cfg.Run(); err != nil {
		fail("Error: %v", err)
	}
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

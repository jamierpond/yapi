// Package imageprinter provides inline terminal image printing.
package imageprinter

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/blacktop/go-termimg"
)

// Config controls how images are rendered.
type Config struct {
	MaxWidth  int // 0 = use default
	MaxHeight int // 0 = use default
}

// Default size constraints (in terminal cells)
const (
	DefaultMaxWidth  = 80
	DefaultMaxHeight = 30
)

// inITerm2 returns true if running in iTerm2.
func inITerm2() bool {
	return os.Getenv("ITERM_SESSION_ID") != "" || os.Getenv("LC_TERMINAL") == "iTerm2"
}

// inKitty returns true if running in Kitty.
func inKitty() bool {
	return os.Getenv("KITTY_WINDOW_ID") != ""
}

// inGhostty returns true if running in Ghostty.
func inGhostty() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty" || os.Getenv("GHOSTTY_RESOURCES_DIR") != ""
}

// Print renders an image from raw bytes to stdout.
func Print(data []byte, cfg Config) error {
	img, err := termimg.From(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	w := cfg.MaxWidth
	if w == 0 {
		w = DefaultMaxWidth
	}
	h := cfg.MaxHeight
	if h == 0 {
		h = DefaultMaxHeight
	}

	img = img.Width(w).Height(h).Scale(termimg.ScaleFit)

	// Force protocol based on terminal
	switch {
	case inKitty(), inGhostty():
		img = img.Protocol(termimg.Kitty)
	case inITerm2():
		img = img.Protocol(termimg.ITerm2)
	}

	return img.Print()
}

// PrintFile renders an image from a file path to stdout.
func PrintFile(path string, cfg Config) error {
	img, err := termimg.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	w := cfg.MaxWidth
	if w == 0 {
		w = DefaultMaxWidth
	}
	h := cfg.MaxHeight
	if h == 0 {
		h = DefaultMaxHeight
	}

	img = img.Width(w).Height(h).Scale(termimg.ScaleFit)

	switch {
	case inKitty(), inGhostty():
		img = img.Protocol(termimg.Kitty)
	case inITerm2():
		img = img.Protocol(termimg.ITerm2)
	}

	return img.Print()
}

// IsSupported returns true if the terminal likely supports image rendering.
func IsSupported() bool {
	return inKitty() || inGhostty() || inITerm2()
}

// IsImageContentType returns true if the content type indicates an image.
func IsImageContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "image/")
}

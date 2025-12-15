// Package imageprinter provides inline terminal image printing.
package imageprinter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blacktop/go-termimg"
)

// Config controls how images are rendered.
type Config struct {
	MaxWidth  int // 0 = fill terminal
	MaxHeight int // 0 = fill terminal
}

// Print renders an image from raw bytes to stdout.
func Print(data []byte, cfg Config) error {
	img, err := termimg.From(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply size constraints if specified
	if cfg.MaxWidth > 0 {
		img = img.Width(cfg.MaxWidth)
	}
	if cfg.MaxHeight > 0 {
		img = img.Height(cfg.MaxHeight)
	}

	// Scale to fit within bounds while preserving aspect ratio
	img = img.Scale(termimg.ScaleFit)

	return img.Print()
}

// PrintFile renders an image from a file path to stdout.
func PrintFile(path string, cfg Config) error {
	img, err := termimg.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}

	if cfg.MaxWidth > 0 {
		img = img.Width(cfg.MaxWidth)
	}
	if cfg.MaxHeight > 0 {
		img = img.Height(cfg.MaxHeight)
	}

	img = img.Scale(termimg.ScaleFit)

	return img.Print()
}

// PrintReader renders an image from a reader to stdout.
func PrintReader(r io.Reader, cfg Config) error {
	img, err := termimg.From(r)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	if cfg.MaxWidth > 0 {
		img = img.Width(cfg.MaxWidth)
	}
	if cfg.MaxHeight > 0 {
		img = img.Height(cfg.MaxHeight)
	}

	img = img.Scale(termimg.ScaleFit)

	return img.Print()
}

// IsSupported returns true if the terminal supports any image protocol.
func IsSupported() bool {
	// Check for known terminal image protocol support
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	if os.Getenv("ITERM_SESSION_ID") != "" || os.Getenv("LC_TERMINAL") == "iTerm2" {
		return true
	}
	// Halfblocks fallback always works in any terminal
	return true
}

// IsImageContentType returns true if the content type indicates an image.
func IsImageContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "image/")
}

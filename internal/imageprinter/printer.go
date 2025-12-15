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
	MaxWidth  int  // 0 = use default
	MaxHeight int  // 0 = use default
	Dither    bool // Enable dithering (useful for halfblocks/sixel)
}

// Default size constraints (in terminal cells)
const (
	DefaultMaxWidth  = 80
	DefaultMaxHeight = 30
)

// Print renders an image from raw bytes to stdout.
// Uses go-termimg's auto-detection to find the best protocol:
// Kitty > iTerm2 > Sixel > Halfblocks (universal fallback)
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

	if cfg.Dither {
		img = img.Dither(true)
	}

	// Let go-termimg auto-detect the best protocol
	// This supports: Kitty, iTerm2, Sixel, and Halfblocks (universal fallback)
	img = img.Protocol(termimg.Auto)

	// Debug: show which protocol is being used
	if os.Getenv("YAPI_DEBUG_IMG") != "" {
		protocols := termimg.DetermineProtocols()
		detected := termimg.DetectProtocol()
		fmt.Fprintf(os.Stderr, "[imageprinter] available_protocols=%v detected=%s\n", protocols, detected)
		renderer, err := img.GetRenderer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[imageprinter] GetRenderer error: %v\n", err)
		} else if renderer != nil {
			fmt.Fprintf(os.Stderr, "[imageprinter] using_protocol=%s\n", renderer.Protocol())
		}
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

	if cfg.Dither {
		img = img.Dither(true)
	}

	// Let go-termimg auto-detect the best protocol
	img = img.Protocol(termimg.Auto)

	return img.Print()
}

// IsSupported returns true if the terminal supports image rendering.
// With halfblocks fallback, this is always true - images can be displayed
// in any terminal that supports Unicode.
func IsSupported() bool {
	// Halfblocks is always available as a fallback, so we always support images
	return true
}

// DetectedProtocol returns the protocol that will be used for image rendering.
func DetectedProtocol() string {
	return termimg.DetectProtocol().String()
}

// AvailableProtocols returns all protocols supported by the current terminal.
func AvailableProtocols() []string {
	protocols := termimg.DetermineProtocols()
	result := make([]string, len(protocols))
	for i, p := range protocols {
		result[i] = p.String()
	}
	return result
}

// IsImageContentType returns true if the content type indicates an image.
func IsImageContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "image/")
}

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

// inTmux returns true if running inside tmux.
func inTmux() bool {
	return os.Getenv("TMUX") != ""
}

// wrapTmuxPassthrough wraps escape sequences for tmux DCS passthrough.
// Tmux requires escapes to be doubled inside the passthrough sequence.
// See: https://github.com/tmux/tmux/issues/1388
func wrapTmuxPassthrough(data []byte) []byte {
	// Double all escape characters (\x1b -> \x1b\x1b)
	escaped := bytes.ReplaceAll(data, []byte{0x1b}, []byte{0x1b, 0x1b})

	var buf bytes.Buffer
	// DCS tmux; prefix
	buf.Write([]byte{0x1b, 'P', 't', 'm', 'u', 'x', ';'})
	buf.Write(escaped)
	// String terminator
	buf.Write([]byte{0x1b, '\\'})

	return buf.Bytes()
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

	// Force iTerm2 protocol when in iTerm2 - Kitty placement is experimental
	// and causes cursor/prompt issues especially in tmux
	if inITerm2() {
		img = img.Protocol(termimg.ITerm2)
	}

	// When inside tmux, buffer output and wrap with DCS passthrough
	// to avoid timeout issues. See: https://github.com/tmux/tmux/issues/1388
	if inTmux() {
		return printWithTmuxPassthrough(img)
	}

	return img.Print()
}

// printWithTmuxPassthrough captures image output and wraps it for tmux DCS passthrough.
// This buffers the entire output before sending to avoid tmux's DCS timeout.
func printWithTmuxPassthrough(img *termimg.Image) error {
	// Capture stdout to a buffer
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	os.Stdout = w

	// Print to the pipe
	printErr := img.Print()

	// Restore stdout and close write end
	w.Close()
	os.Stdout = oldStdout

	// Read all captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return fmt.Errorf("failed to read output: %w", err)
	}
	r.Close()

	if printErr != nil {
		return printErr
	}

	// Wrap with tmux passthrough and write all at once
	wrapped := wrapTmuxPassthrough(buf.Bytes())
	_, err = os.Stdout.Write(wrapped)
	return err
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

	if inITerm2() {
		img = img.Protocol(termimg.ITerm2)
	}

	// When inside tmux, buffer output and wrap with DCS passthrough
	if inTmux() {
		return printWithTmuxPassthrough(img)
	}

	return img.Print()
}

// IsSupported returns true if the terminal supports any image protocol.
func IsSupported() bool {
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	if inITerm2() {
		return true
	}
	return true
}

// IsImageContentType returns true if the content type indicates an image.
func IsImageContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "image/")
}

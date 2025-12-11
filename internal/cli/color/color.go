// Package color provides CLI color constants matching the yapi theme.
// Uses ANSI true color (24-bit) for exact RGB color matching.
package color

import "fmt"

// Theme RGB values from internal/tui/theme/theme.go
const (
	// Accent: #ff9e64
	accentR, accentG, accentB = 255, 158, 100
	// Green: #69DB7C
	greenR, greenG, greenB = 105, 219, 124
	// Red: #FF6B6B
	redR, redG, redB = 255, 107, 107
	// Yellow: #FFE066
	yellowR, yellowG, yellowB = 255, 224, 102
	// Cyan/Info: #7aa2f7
	cyanR, cyanG, cyanB = 122, 162, 247
)

// ANSI true color escape sequences
var (
	accentSeq   = fmt.Sprintf("\033[38;2;%d;%d;%dm", accentR, accentG, accentB)
	greenSeq    = fmt.Sprintf("\033[38;2;%d;%d;%dm", greenR, greenG, greenB)
	redSeq      = fmt.Sprintf("\033[38;2;%d;%d;%dm", redR, redG, redB)
	yellowSeq   = fmt.Sprintf("\033[38;2;%d;%d;%dm", yellowR, yellowG, yellowB)
	cyanSeq     = fmt.Sprintf("\033[38;2;%d;%d;%dm", cyanR, cyanG, cyanB)
	dimSeq      = "\033[2m"
	boldSeq     = "\033[1m"
	resetSeq    = "\033[0m"
	accentBgSeq = fmt.Sprintf("\033[48;2;%d;%d;%dm\033[38;2;255;255;255m", accentR, accentG, accentB) // white on orange
)

// noColor tracks if color output is disabled
var noColor bool

// SetNoColor globally disables color output.
func SetNoColor(disabled bool) {
	noColor = disabled
}

// wrap returns text wrapped in color sequence, respecting NoColor
func wrap(seq, text string) string {
	if noColor {
		return text
	}
	return seq + text + resetSeq
}

// Accent formats text with the theme accent color (#ff9e64 orange)
func Accent(text string) string {
	return wrap(accentSeq, text)
}

// Green formats text with theme green (#69DB7C)
func Green(text string) string {
	return wrap(greenSeq, text)
}

// Red formats text with theme red (#FF6B6B)
func Red(text string) string {
	return wrap(redSeq, text)
}

// Yellow formats text with theme yellow (#FFE066)
func Yellow(text string) string {
	return wrap(yellowSeq, text)
}

// Cyan formats text with theme cyan/info (#7aa2f7)
func Cyan(text string) string {
	return wrap(cyanSeq, text)
}

// Dim formats text with faint/dim style
func Dim(text string) string {
	return wrap(dimSeq, text)
}

// AccentBg formats text with white on accent background
func AccentBg(text string) string {
	return wrap(accentBgSeq, text)
}

// Bold formats text with bold style
func Bold(text string) string {
	return wrap(boldSeq, text)
}

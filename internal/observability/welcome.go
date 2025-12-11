package observability

import (
	"fmt"
	"os"

	"yapi.run/cli/internal/cli/color"
)

// RunWelcome displays a non-blocking first-run banner (Go toolchain style).
// Does not prompt for input - defaults to local collection mode.
func RunWelcome() {
	if !IsFirstRun() {
		return
	}

	// Initialize with "local" default immediately (Go vibe: safe default)
	_ = SetMode(ModeLocal)

	// Non-interactive or CI: skip the banner entirely
	if !isInteractive() || os.Getenv("CI") != "" {
		return
	}

	// Print a non-intrusive banner (Go style)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.Dim("yapi: telemetry is enabled (local only)"))
	fmt.Fprintln(os.Stderr, color.Dim("yapi: to help improve yapi, run: ")+color.Bold("yapi telemetry on"))
	fmt.Fprintln(os.Stderr)
}

// isInteractive returns true if stdin is a terminal
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

package observability

import (
	"fmt"
	"os"

	"yapi.run/cli/internal/cli/color"
)

// RunWelcome initializes telemetry mode on first run.
// Does not prompt or print anything - just sets the safe default.
func RunWelcome() {
	if !IsFirstRun() {
		return
	}

	// Initialize with "local" default immediately (Go vibe: safe default)
	_ = SetMode(ModeLocal)
}

// PrintFirstRunBanner prints a non-blocking banner at the END of command execution.
// Only prints on first run, in interactive mode, outside CI.
func PrintFirstRunBanner() {
	// Only show if mode was just set (first run indicator)
	cfg, err := LoadUserConfig()
	if err != nil {
		return
	}

	// Check if this is a fresh "local" mode (first run state)
	if cfg.TelemetryMode != ModeLocal {
		return
	}

	// Check for the marker file that indicates we've shown the banner
	if !shouldShowBanner() {
		return
	}

	// Non-interactive or CI: skip the banner
	if !isInteractive() || os.Getenv("CI") != "" {
		return
	}

	// Mark that we've shown the banner
	markBannerShown()

	// Print banner at the end in cyan (blue)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.Cyan("yapi writes (anonymous) usage stats to ")+color.Bold("~/"+LogFileName)+color.Cyan(" (nothing is uploaded)."))
	fmt.Fprintln(os.Stderr, color.Cyan("Check out the file - if you're comfortable sharing, run ")+color.Bold("yapi telemetry on"))
	fmt.Fprintln(os.Stderr, color.Cyan("You can change this later with ")+color.Bold("yapi telemetry off"))
	fmt.Fprintln(os.Stderr, color.Dim(""))
	fmt.Fprintln(os.Stderr, color.Dim("I (jamierpond@github.com) am building yapi in public"))
	fmt.Fprintln(os.Stderr, color.Dim("I just wanna see people using a thing I made :)"))
}

// shouldShowBanner returns true if we haven't shown the banner yet
func shouldShowBanner() bool {
	path, err := bannerMarkerPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return os.IsNotExist(err)
}

// markBannerShown creates a marker file so we don't show the banner again
func markBannerShown() {
	path, err := bannerMarkerPath()
	if err != nil {
		return
	}
	_ = os.WriteFile(path, []byte{}, 0644)
}

// bannerMarkerPath returns the path to the banner shown marker
func bannerMarkerPath() (string, error) {
	dir, err := yapiConfigDir()
	if err != nil {
		return "", err
	}
	return dir + "/.banner_shown", nil
}

// isInteractive returns true if stdin is a terminal
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

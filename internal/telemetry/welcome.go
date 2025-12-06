package telemetry

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RunWelcome displays the first-run welcome message and asks for telemetry consent.
// Returns true if telemetry was enabled, false otherwise.
// Only runs if this is the first run (no preference set yet).
func RunWelcome() bool {
	if !IsFirstRun() {
		return false
	}

	// Check if we're in an interactive terminal
	if !isInteractive() {
		// Non-interactive: default to disabled, don't save preference
		return false
	}

	fmt.Println()
	fmt.Println("  Welcome to yapi!")
	fmt.Println()
	fmt.Println("  Hey Jamie here! Thank you SO MUCH for trying out yapi!")
	fmt.Println("  I'd like to ask a HUGE favour and to enable anonymous analytics")
	fmt.Println("  No personal data, request contents, or URLs are EVER collected.")
	fmt.Println()
	fmt.Println("  I want to make yapi as awesome and useful as I can, so seeing")
	fmt.Println("  how real people are using it in the world would be super useful")
	fmt.Println()
	fmt.Println("  You can change this anytime:")
	fmt.Println("    - Set YAPI_NO_ANALYTICS=1 in your environment")
	fmt.Println("    - Edit ~/.config/yapi/config.json")
	fmt.Println()
	fmt.Print("  Enable anonymous analytics? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// On error, default to disabled
		_ = SetTelemetryEnabled(false)
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))

	// Default to yes (empty input or "y" or "yes")
	enabled := input == "" || input == "y" || input == "yes"

	if err := SetTelemetryEnabled(enabled); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not save preference: %v\n", err)
	}

	if enabled {
		fmt.Println("  Thanks! Analytics enabled.")
	} else {
		fmt.Println("  Analytics disabled.")
	}
	fmt.Println()

	return enabled
}

// isInteractive returns true if stdin is a terminal
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

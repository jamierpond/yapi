package observability

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"yapi.run/cli/internal/cli/color"
)

// RunWelcome displays the first-run welcome message and asks for observability consent.
// Returns true if observability was enabled, false otherwise.
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
	fmt.Println(color.AccentBg(" yapi "))
	fmt.Println()
	fmt.Println(color.Cyan("  Hey Jamie here! Thank you ") + color.Green("SO MUCH") + color.Cyan(" for trying out yapi!"))
	fmt.Println("  I'd like to ask a HUGE favour and to enable " + color.Yellow("local logging"))
	fmt.Println(color.Dim("  Just which commands you ran, which features you used, and if you hit any errors."))
	fmt.Println(color.Dim("  No personal data, request contents, or URLs are EVER collected."))
	fmt.Println(color.Dim("  Logs are stored locally at ~/yapi-log.txt"))
	fmt.Println()
	fmt.Println("  I want to make yapi as awesome and useful as I can, so seeing")
	fmt.Println("  how real people are using it in the world would be super useful")
	fmt.Println()
	fmt.Println(color.Dim("  You can change this anytime:"))
	fmt.Println(color.Dim("    - Set YAPI_NO_ANALYTICS=1 in your environment"))
	fmt.Println(color.Dim("    - Edit ~/.config/yapi/config.json"))
	fmt.Println()
	fmt.Print("  Enable logging? " + color.Dim("[y/N]: "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// On error, default to disabled
		_ = SetObservabilityEnabled(false)
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))

	// Default to no (only enable on explicit "y" or "yes")
	enabled := input == "y" || input == "yes"

	if err := SetObservabilityEnabled(enabled); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not save preference: %v\n", err)
	}

	fmt.Println()
	if enabled {
		fmt.Println(color.Green("  Logging enabled. Thank you!"))
	} else {
		fmt.Println(color.Dim("  Logging disabled."))
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

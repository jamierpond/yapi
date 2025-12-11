package observability

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"yapi.run/cli/internal/cli/color"
)

// RunWelcome displays the first-run welcome message and asks for observability consent.
// Only runs if this is the first run (no preference set yet).
func RunWelcome() {
	if !IsFirstRun() {
		return
	}

	// Check if we're in an interactive terminal
	if !isInteractive() {
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(color.AccentBg(" yapi "))
	fmt.Println()
	fmt.Println(color.Cyan("  Hey Jamie here! Thank you ") + color.Green("SO MUCH") + color.Cyan(" for trying out yapi!"))
	fmt.Println()
	fmt.Println("  I'd love to understand how people use yapi so I can make it better.")
	fmt.Println(color.Dim("  Just which commands you ran, which features you used, and if you hit any errors."))
	fmt.Println(color.Dim("  No personal data, request contents, or URLs are EVER collected."))
	fmt.Println()
	fmt.Println("  Everything is logged locally to " + color.Yellow("~/yapi-log.txt") + " first.")
	fmt.Println("  You can audit exactly what would be sent before enabling analytics.")
	fmt.Println()

	// Ask about local file logging (default yes)
	fmt.Print("  Enable local logging? " + color.Dim("[Y/n]: "))

	input, err := reader.ReadString('\n')
	if err != nil {
		_ = SetFileLoggingEnabled(true)
		_ = SetPosthogEnabled(false)
		return
	}
	input = strings.TrimSpace(strings.ToLower(input))
	fileLogging := input != "n" && input != "no"
	_ = SetFileLoggingEnabled(fileLogging)

	fmt.Println()

	// Ask about PostHog - emphasize it's the same data
	fmt.Println("  Send the " + color.Yellow("same data") + " to PostHog so I can see aggregate patterns?")
	fmt.Println(color.Dim("  (100% parity with ~/yapi-log.txt - nothing extra)"))
	fmt.Print("  Enable anonymous analytics? " + color.Dim("[y/N]: "))

	input, err = reader.ReadString('\n')
	if err != nil {
		_ = SetPosthogEnabled(false)
		return
	}
	posthog := strings.TrimSpace(strings.ToLower(input)) == "y" || strings.TrimSpace(strings.ToLower(input)) == "yes"
	_ = SetPosthogEnabled(posthog)

	fmt.Println()

	// Summary
	if fileLogging || posthog {
		fmt.Println(color.Green("  Thank you!"))
		if fileLogging {
			fmt.Println(color.Dim("    - Local logging: enabled"))
		}
		if posthog {
			fmt.Println(color.Dim("    - Anonymous analytics: enabled"))
		}
	} else {
		fmt.Println(color.Dim("  No worries, observability disabled."))
	}

	fmt.Println()
	fmt.Println(color.Dim("  You can change this anytime:"))
	fmt.Println(color.Dim("    - Set YAPI_NO_ANALYTICS=1 to disable all"))
	fmt.Println(color.Dim("    - Edit ~/.config/yapi/config.json"))
	fmt.Println()
}

// isInteractive returns true if stdin is a terminal
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

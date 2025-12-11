package observability

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"yapi.run/cli/internal/cli/color"
)

// RunWelcome displays the first-run welcome message.
func RunWelcome() {
	if !IsFirstRun() {
		return
	}

	// If non-interactive or CI, default to disabled and don't block
	if !isInteractive() || os.Getenv("CI") != "" {
		_ = SetFileLoggingEnabled(false)
		_ = SetPosthogEnabled(false)
		return
	}

	fmt.Println()
	fmt.Println(color.AccentBg(" yapi "))
	fmt.Println()
	fmt.Println(color.Bold("  Thanks for installing yapi!"))
	fmt.Println()
	fmt.Println("  I'm building this in public. To help me improve it, yapi collects")
	fmt.Println("  anonymous usage stats (commands run, errors hit).")
	fmt.Println()
	fmt.Println(color.Dim("  - No personal data, IPs, or request bodies are ever sent."))
	fmt.Println(color.Dim("  - You can audit everything sent at ~/yapi-log.txt"))
	fmt.Println()

	fmt.Print(color.Green("  OK to send anonymous stats? ") + color.Dim("[Y/n] "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	// Default to YES if error or enter
	confirmed := true
	if err == nil {
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "n" || input == "no" {
			confirmed = false
		}
	}

	// Always enable file logging so they can audit
	_ = SetFileLoggingEnabled(true)
	_ = SetPosthogEnabled(confirmed)

	fmt.Println()
	if confirmed {
		fmt.Println(color.Dim("  Awesome, thanks! (Disable anytime via `yapi telemetry disable`)"))
	} else {
		fmt.Println(color.Dim("  Understood. Telemetry disabled."))
	}
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

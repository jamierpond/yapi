// Package middleware provides Cobra command middleware for the yapi CLI.
package middleware

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"yapi.run/cli/internal/observability"
)

// WrapWithObservability recursively wraps all commands with observability instrumentation.
// This automatically captures command name, flags used (not values), args count, timing, and success/failure.
func WrapWithObservability(cmd *cobra.Command) {
	// Recursively wrap all child commands first
	for _, c := range cmd.Commands() {
		WrapWithObservability(c)
	}

	// Skip noise commands (completion generates many internal calls)
	if cmd.Name() == "completion" || cmd.Name() == "__complete" {
		return
	}

	// Get the original run function
	originalRunE := cmd.RunE
	if originalRunE == nil && cmd.Run != nil {
		originalRun := cmd.Run
		originalRunE = func(c *cobra.Command, args []string) error {
			originalRun(c, args)
			return nil
		}
	}

	// If no run function, nothing to wrap
	if originalRunE == nil {
		return
	}

	// Clear the Run field since we're using RunE
	cmd.Run = nil

	// Wrap with observability
	cmd.RunE = func(c *cobra.Command, args []string) error {
		start := time.Now()

		// Collect properties from flags
		props := make(map[string]any)

		// Go vibe: Only record that the flag was used, NOT its value
		// This avoids capturing sensitive data like URLs, tokens, etc.
		cmd.Flags().Visit(func(f *pflag.Flag) {
			props["flag_used_"+f.Name] = true
		})

		// Only record args count, not the args themselves
		props["args_count"] = len(args)

		// Execute the original command
		err := originalRunE(c, args)

		// Track command execution
		props["duration_ms"] = time.Since(start).Milliseconds()
		props["success"] = err == nil
		if err != nil {
			// Record error type, not the full message (which may contain sensitive paths)
			props["has_error"] = true
		}
		observability.Track("cmd_"+cmd.Name(), props)

		return err
	}
}

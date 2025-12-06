// Package middleware provides Cobra command middleware for the yapi CLI.
package middleware

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"yapi.run/cli/internal/telemetry"
)

// WrapWithTelemetry recursively wraps all commands with telemetry instrumentation.
// This automatically captures command name, flags, args, timing, and success/failure.
func WrapWithTelemetry(cmd *cobra.Command) {
	// Recursively wrap all child commands first
	for _, c := range cmd.Commands() {
		WrapWithTelemetry(c)
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

	// Wrap with telemetry
	cmd.RunE = func(c *cobra.Command, args []string) error {
		start := time.Now()

		// Collect properties from flags
		props := make(map[string]interface{})
		cmd.Flags().Visit(func(f *pflag.Flag) {
			props["flag_"+f.Name] = f.Value.String()
		})
		props["args_count"] = len(args)

		// Execute the original command
		err := originalRunE(c, args)

		// Track command execution
		props["duration_ms"] = time.Since(start).Milliseconds()
		props["success"] = err == nil
		if err != nil {
			props["error"] = err.Error()
		}
		telemetry.Track("cmd_"+cmd.Name(), props)

		return err
	}
}

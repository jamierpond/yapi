package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/share"
	"yapi.run/cli/internal/validation"
)

func shareE(cmd *cobra.Command, args []string) error {
	filename, _, err := selectConfigFile(args, "share")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filename) //nolint:gosec // user-provided file path
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// Validate the config
	analysis, analysisErr := validation.Analyze(content, validation.AnalyzeOptions{FilePath: filename})
	if analysisErr != nil {
		return fmt.Errorf("failed to analyze config: %w", analysisErr)
	}
	hasErrors := analysis != nil && analysis.HasErrors()
	hasWarnings := analysis != nil && len(analysis.Warnings) > 0

	encoded, err := share.Encode(content)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	url := "https://yapi.run/c/" + encoded

	// Stats
	originalSize := len(data)
	compressedSize := len(encoded)
	ratio := float64(compressedSize) / float64(originalSize) * 100
	lines := strings.Count(content, "\n") + 1

	// Fancy output to stderr
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.AccentBg(" yapi share "))
	fmt.Fprintln(os.Stderr)

	//nolint:gocritic // ifElseChain: switch not suitable for boolean conditions
	if hasErrors {
		fmt.Fprintln(os.Stderr, "  "+color.Yellow("Heads up: this yap has validation errors!"))
		fmt.Fprintln(os.Stderr)
		for _, d := range analysis.Diagnostics {
			if d.Severity == validation.SeverityError {
				fmt.Fprintln(os.Stderr, "  "+color.Red(d.Message))
			}
		}
		fmt.Fprintln(os.Stderr)
	} else if hasWarnings {
		fmt.Fprintln(os.Stderr, "  "+color.Yellow("Your yap has warnings, but it's ready to share!"))
	} else {
		fmt.Fprintln(os.Stderr, "  "+color.Green("Your yap is ready to share!"))
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.Dim("  file     ")+filepath.Base(filename))
	fmt.Fprintln(os.Stderr, color.Dim("  lines    ")+fmt.Sprintf("%d", lines))
	fmt.Fprintln(os.Stderr, color.Dim("  size     ")+fmt.Sprintf("%s -> %s (%.0f%%)", formatBytes(originalSize), formatBytes(compressedSize), ratio))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  "+color.Cyan(url))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.Dim("  The entire request is encoded in the URL - just share it!"))
	fmt.Fprintln(os.Stderr)

	// Only print raw URL to stdout when piping (not a terminal)
	if stat, _ := os.Stdout.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Println(url)
	}
	return nil
}

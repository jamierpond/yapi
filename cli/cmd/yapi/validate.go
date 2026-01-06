package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/validation"
)

func validateE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	all, _ := cmd.Flags().GetBool("all")

	// Handle --all flag
	if all {
		return validateAllFiles(args, jsonOutput)
	}

	var path string
	var err error

	// If no file provided, look for project config first
	if len(args) == 0 {
		cwd, _ := os.Getwd()
		if projectRoot, findErr := config.FindProjectRoot(cwd); findErr == nil {
			// Found a project config, validate it
			configPath := filepath.Join(projectRoot, "yapi.config.yml")
			if _, statErr := os.Stat(configPath); statErr == nil {
				return validateProjectConfigFile(configPath, jsonOutput)
			}
			configPath = filepath.Join(projectRoot, "yapi.config.yaml")
			if _, statErr := os.Stat(configPath); statErr == nil {
				return validateProjectConfigFile(configPath, jsonOutput)
			}
		}
	}

	// Otherwise use normal file selection (including project config files)
	path, _, err = selectConfigFileIncludingProject(args, "validate")
	if err != nil {
		if jsonOutput {
			outputValidateError(err)
			return nil
		}
		return err
	}

	// Check if this is a project config file
	fileName := filepath.Base(path)
	if fileName == "yapi.config.yml" || fileName == "yapi.config.yaml" {
		return validateProjectConfigFile(path, jsonOutput)
	}

	data, err := utils.ReadInput(path)
	if err != nil {
		if jsonOutput {
			outputValidateError(err)
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	analysis, err := validation.Analyze(string(data), validation.AnalyzeOptions{FilePath: path})
	if err != nil {
		if jsonOutput {
			outputValidateError(err)
			return nil
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(analysis.ToJSON())
		return nil
	}

	return outputValidateText(analysis, path, data)
}

func validateProjectConfigFile(path string, jsonOutput bool) error {
	// Try to load the project config
	projectRoot := filepath.Dir(path)
	_, err := config.LoadProject(projectRoot)

	if jsonOutput {
		if err != nil {
			out := validation.JSONOutput{
				Valid: false,
				Diagnostics: []validation.JSONDiagnostic{{
					Severity: "error",
					Message:  fmt.Sprintf("Invalid project config: %v", err),
					Line:     0,
					Col:      0,
				}},
				Warnings: []string{},
			}
			_ = json.NewEncoder(os.Stdout).Encode(out)
		} else {
			out := validation.JSONOutput{
				Valid:       true,
				Diagnostics: []validation.JSONDiagnostic{},
				Warnings:    []string{},
			}
			_ = json.NewEncoder(os.Stdout).Encode(out)
		}
		return nil
	}

	// Text output
	data, readErr := os.ReadFile(path) // #nosec G304 -- path is validated user-provided config file path
	if readErr != nil {
		return fmt.Errorf("failed to read config: %w", readErr)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.AccentBg(" yapi validate "))
	fmt.Fprintln(os.Stderr)

	absPath, _ := filepath.Abs(path)
	fmt.Fprintln(os.Stderr, "  "+color.Dim("file     ")+filepath.Base(absPath))
	if dir := filepath.Dir(absPath); dir != "" && dir != "." {
		fmt.Fprintln(os.Stderr, "  "+color.Dim("path     ")+dir)
	}

	lines := strings.Count(string(data), "\n") + 1
	size := len(data)
	fmt.Fprintln(os.Stderr, "  "+color.Dim("lines    ")+fmt.Sprintf("%d", lines))
	fmt.Fprintln(os.Stderr, "  "+color.Dim("size     ")+formatBytes(size))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		fmt.Fprintln(os.Stderr, color.Red("[ERROR] ")+err.Error())
		fmt.Fprintln(os.Stderr)
		return errors.New("validation errors")
	}

	fmt.Fprintln(os.Stderr, "  "+color.Green("Valid project configuration!"))
	fmt.Fprintln(os.Stderr)
	return nil
}

func outputValidateError(err error) {
	out := validation.JSONOutput{
		Valid: false,
		Diagnostics: []validation.JSONDiagnostic{{
			Severity: "error",
			Message:  err.Error(),
			Line:     0,
			Col:      0,
		}},
		Warnings: []string{},
	}
	_ = json.NewEncoder(os.Stdout).Encode(out)
}

func outputValidateText(analysis *validation.Analysis, path string, data []byte) error {
	hasOutput := len(analysis.Warnings) > 0 || len(analysis.Diagnostics) > 0

	// Print file info header
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.AccentBg(" yapi validate "))
	fmt.Fprintln(os.Stderr)

	// Show file path (or stdin indicator)
	if path == "-" {
		fmt.Fprintln(os.Stderr, "  "+color.Dim("source   stdin"))
	} else {
		absPath, _ := filepath.Abs(path)
		fmt.Fprintln(os.Stderr, "  "+color.Dim("file     ")+filepath.Base(absPath))
		if dir := filepath.Dir(absPath); dir != "" && dir != "." {
			fmt.Fprintln(os.Stderr, "  "+color.Dim("path     ")+dir)
		}
	}

	// Show file stats
	lines := strings.Count(string(data), "\n") + 1
	size := len(data)
	fmt.Fprintln(os.Stderr, "  "+color.Dim("lines    ")+fmt.Sprintf("%d", lines))
	fmt.Fprintln(os.Stderr, "  "+color.Dim("size     ")+formatBytes(size))
	fmt.Fprintln(os.Stderr)

	if hasOutput {
		// Print errors and warnings
		validation.PrintErrors(analysis, os.Stderr, false)
		validation.PrintWarnings(analysis, os.Stderr, false)
		fmt.Fprintln(os.Stderr)
	} else {
		fmt.Fprintln(os.Stderr, "  "+color.Green("Valid!"))
		fmt.Fprintln(os.Stderr)
	}

	if analysis.HasErrors() {
		return errors.New("validation errors")
	}
	return nil
}

// validateAllFiles validates all yapi files in a directory
func validateAllFiles(args []string, jsonOutput bool) error {
	// Determine search directory
	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	// Find all yapi files
	yapiFiles, err := findAllYapiFiles(searchDir)
	if err != nil {
		return fmt.Errorf("failed to find yapi files: %w", err)
	}

	if len(yapiFiles) == 0 {
		if jsonOutput {
			// Output empty JSON array
			fmt.Println("[]")
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.yapi, *.yapi.yml, or *.yapi.yaml files found"))
		}
		return nil
	}

	if !jsonOutput {
		fmt.Fprintf(os.Stderr, "%s\n\n", color.Accent(fmt.Sprintf("Validating %d file(s)...", len(yapiFiles))))
	}

	// Validate each file
	type validationResult struct {
		file     string
		valid    bool
		analysis *validation.Analysis
		err      error
	}

	var results []validationResult
	validCount := 0

	for _, filePath := range yapiFiles {
		relPath, _ := filepath.Rel(searchDir, filePath)

		// Read file
		data, err := os.ReadFile(filePath) // #nosec G304 -- filePath is from filesystem walk
		if err != nil {
			results = append(results, validationResult{
				file:  relPath,
				valid: false,
				err:   err,
			})
			continue
		}

		// Validate
		analysis, err := validation.Analyze(string(data), validation.AnalyzeOptions{FilePath: filePath})
		if err != nil {
			results = append(results, validationResult{
				file:  relPath,
				valid: false,
				err:   err,
			})
			continue
		}

		valid := !analysis.HasErrors()
		if valid {
			validCount++
		}

		results = append(results, validationResult{
			file:     relPath,
			valid:    valid,
			analysis: analysis,
		})

		if !jsonOutput {
			if valid {
				fmt.Fprintf(os.Stderr, "%s %s\n", color.Green("OK"), relPath)
			} else {
				fmt.Fprintf(os.Stderr, "%s %s\n", color.Red("X"), relPath)
			}
		}
	}

	if jsonOutput {
		// Output JSON array of results
		type jsonResult struct {
			File        string                      `json:"file"`
			Valid       bool                        `json:"valid"`
			Diagnostics []validation.JSONDiagnostic `json:"diagnostics,omitempty"`
			Error       string                      `json:"error,omitempty"`
		}

		jsonResults := make([]jsonResult, len(results))
		for i, r := range results {
			result := jsonResult{
				File:  r.file,
				Valid: r.valid,
			}
			if r.err != nil {
				result.Error = r.err.Error()
			} else if r.analysis != nil {
				result.Diagnostics = r.analysis.ToJSON().Diagnostics
			}
			jsonResults[i] = result
		}

		return json.NewEncoder(os.Stdout).Encode(jsonResults)
	}

	// Text output - print summary
	fmt.Fprintf(os.Stderr, "\n")
	if validCount == len(results) {
		fmt.Fprintf(os.Stderr, "%s\n", color.Green(fmt.Sprintf("All %d file(s) are valid", validCount)))
		return nil
	}

	invalidCount := len(results) - validCount
	fmt.Fprintf(os.Stderr, "%s\n", color.Red(fmt.Sprintf("%d of %d file(s) have errors", invalidCount, len(results))))

	// List files with errors
	fmt.Fprintf(os.Stderr, "\n%s\n", color.Red("Files with errors:"))
	for _, r := range results {
		if !r.valid {
			fmt.Fprintf(os.Stderr, "  %s %s\n", color.Red("X"), r.file)
			if r.err != nil {
				fmt.Fprintf(os.Stderr, "    %s\n", color.Dim(r.err.Error()))
			} else if r.analysis != nil && len(r.analysis.Diagnostics) > 0 {
				for _, d := range r.analysis.Diagnostics {
					if d.Severity == validation.SeverityError {
						fmt.Fprintf(os.Stderr, "    %s\n", color.Dim(d.Message))
					}
				}
			}
		}
	}

	return fmt.Errorf("%d file(s) have validation errors", invalidCount)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/validation"
)

// runContext holds options for executeRun
type runContext struct {
	path         string
	strict       bool   // If true, return error on failures; if false, print and return nil
	returnErrors bool   // If true, return errors even when strict is false (for stress tests)
	envName      string // Target environment from yapi.config.yml
	jsonOutput   bool   // If true, output structured JSON instead of formatted output
	strictEnv    bool   // If true, error on missing env files and disable OS env fallback
	verbose      bool   // If true, show verbose output (request details, timing, headers)
}

func (app *rootCommand) runInteractiveE(cmd *cobra.Command, args []string) error {
	path, _, err := selectConfigFile(args, "run")
	if err != nil {
		return err
	}
	return app.runConfigPathE(path)
}

func (app *rootCommand) runE(cmd *cobra.Command, args []string) error {
	path, _, err := selectConfigFile(args, "run")
	if err != nil {
		return err
	}

	// Get flags
	envName, _ := cmd.Flags().GetString("env")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	strictEnv, _ := cmd.Flags().GetBool("strict-env")
	verbose, _ := cmd.Flags().GetBool("verbose")
	return app.runConfigPathWithOptionsE(path, envName, jsonOutput, strictEnv, verbose)
}

func (app *rootCommand) watchE(cmd *cobra.Command, args []string) error {
	pretty, _ := cmd.Flags().GetBool("pretty")
	noPretty, _ := cmd.Flags().GetBool("no-pretty")
	envName, _ := cmd.Flags().GetString("env")

	path, fromTUI, err := selectConfigFile(args, "watch")
	if err != nil {
		return err
	}

	usePretty := pretty || (fromTUI && !noPretty)

	if usePretty {
		opts := runner.Options{
			URLOverride:    app.urlOverride,
			NoColor:        app.noColor,
			BinaryOutput:   app.binaryOutput,
			Insecure:       app.insecure,
			ConfigFilePath: path,
		}
		return tui.RunWatch(path, opts)
	}
	return app.watchConfigPath(path, envName)
}

func (app *rootCommand) watchConfigPath(path string, envName string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	clearScreen()
	printWatchHeader(absPath)
	_ = app.executeRunE(runContext{path: absPath, strict: false, envName: envName})

	lastMod, err := getModTime(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		currentMod, err := getModTime(absPath)
		if err != nil {
			// File became inaccessible - print error and continue watching
			fmt.Fprintf(os.Stderr, "%s\n", color.Red("file inaccessible: "+err.Error()))
			continue
		}
		if currentMod != lastMod {
			lastMod = currentMod
			clearScreen()
			printWatchHeader(absPath)
			_ = app.executeRunE(runContext{path: absPath, strict: false, envName: envName})
		}
	}
	return nil
}

func getModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printWatchHeader(path string) {
	fmt.Printf("%s\n\n", color.Accent("yapi watch"))
	fmt.Printf("%s\n", color.Dim("[watching "+filepath.Base(path)+"]"))
	fmt.Printf("%s\n\n", color.Dim("["+time.Now().Format("15:04:05")+"]"))
}

// OutputSavedError is returned when output is too large for terminal and was saved to file.
type OutputSavedError struct {
	Path string
}

func (e *OutputSavedError) Error() string {
	return fmt.Sprintf("output saved to %s (too large for terminal)\nview with: cat %s | jq", e.Path, e.Path)
}

// printResultOptions configures printResult behavior.
type printResultOptions struct {
	skipMeta bool // Don't print URL/Time/Size (already shown in verbose mode)
}

// printResult outputs a single result with optional expectation.
// configPath is used for generating auto-save filenames when output is too large.
// Returns OutputSavedError if output was saved to file instead of printed.
func (app *rootCommand) printResult(result *runner.Result, expectRes *runner.ExpectationResult, configPath string, opts printResultOptions) error {
	var savedPath string
	if result != nil {
		// Check if stdout is a TTY (terminal)
		isTTY := isTerminal(os.Stdout)

		// Check if content is binary
		isBinary := utils.IsBinaryContent(result.Body)

		switch {
		case isBinary && !app.binaryOutput:
			// Skip dumping binary output unless explicitly requested with --binary-output
			if isTTY {
				fmt.Fprintf(os.Stderr, "\n%s\n", color.Yellow("Binary content detected. Output hidden to prevent terminal corruption."))
				fmt.Fprintf(os.Stderr, "%s\n", color.Dim("To display binary output, use --binary-output flag or pipe to a file."))
			}
			// In non-TTY (CI/piped), silently skip binary output
		case result.OutputFile != "":
			// Output was already saved via output_file config - don't write again
			if len(result.Body) > maxOutputSize {
				savedPath = result.OutputFile
			} else {
				// Small enough to print, but also saved to file
				body := output.Highlight(result.Body, result.ContentType, app.noColor)
				fmt.Println(strings.TrimRight(body, "\n\r"))
			}
		default:
			// No output_file specified - render normally (may auto-save if large)
			body := result.Body
			if len(body) <= maxOutputSize {
				body = output.Highlight(body, result.ContentType, app.noColor)
			}
			outputResult := renderOutput(body, configPath)
			savedPath = outputResult.SavedPath
		}

		if !opts.skipMeta {
			printResultMeta(result)
		}
	}
	if expectRes != nil {
		printExpectationResult(expectRes)
	}
	if savedPath != "" {
		return &OutputSavedError{Path: savedPath}
	}
	return nil
}

// executeRunE is the unified execution pipeline for both Run and Watch modes.
// Returns error for middleware to capture.
func (app *rootCommand) executeRunE(ctx runContext) error {
	log := NewLogger(ctx.verbose)

	opts := runner.Options{
		URLOverride:    app.urlOverride,
		NoColor:        app.noColor,
		BinaryOutput:   app.binaryOutput,
		Insecure:       app.insecure,
		Verbose:        ctx.verbose,
		ConfigFilePath: ctx.path,
		StrictEnv:      ctx.strictEnv,
	}

	log.Verbose("Loading config: %s", ctx.path)

	// Load project and environment configuration
	projEnv, err := loadProjectAndEnv(ctx.path, ctx.envName, true)
	if err != nil {
		if ctx.strict || ctx.returnErrors {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s\n", color.Red(err.Error()))
		return nil
	}

	// Apply project settings if found
	if projEnv != nil {
		log.Verbose("Project: %s", projEnv.projectRoot)
		opts.ProjectRoot = projEnv.projectRoot
		if projEnv.envVars != nil {
			log.Verbose("Environment: %s (%d vars)", projEnv.envName, len(projEnv.envVars))
			opts.EnvOverrides = projEnv.envVars
			opts.ProjectEnv = projEnv.envName
		}
	}

	log.Verbose("Sending request...")
	runRes := app.engine.RunConfig(context.Background(), ctx.path, opts)

	// Log response details if available
	if runRes.Result != nil {
		log.Response(runRes.Result.StatusCode, runRes.Result.Headers, runRes.Result.Duration, runRes.Result.BodyBytes)
	} else if runRes.Error != nil {
		log.Verbose("Request failed: %v", runRes.Error)
	}

	// Handle validation/parse errors first
	if runRes.Error != nil && runRes.Analysis == nil {
		if ctx.strict || ctx.returnErrors {
			return runRes.Error
		}
		fmt.Println(color.Red(runRes.Error.Error()))
		return nil
	}

	out, noColor := app.io(ctx.strict)
	validation.PrintErrors(runRes.Analysis, out, noColor)
	if runRes.Analysis != nil && runRes.Analysis.HasErrors() {
		if ctx.strict || ctx.returnErrors {
			return &validation.Error{Diagnostics: runRes.Analysis.Diagnostics}
		}
		return nil
	}

	// Check if this is a chain config
	if runRes.Analysis != nil && len(runRes.Analysis.Chain) > 0 {
		return app.executeChain(ctx, runRes, opts)
	}

	if runRes.Analysis == nil || runRes.Analysis.Request == nil {
		if ctx.strict {
			return errors.New("invalid config")
		}
		return nil
	}

	// Handle JSON output mode
	if ctx.jsonOutput {
		// If we have an error and no result, still output JSON with error info
		if runRes.Result == nil && runRes.Error != nil {
			return output.PrintJSON(output.JSONParams{
				Analysis: runRes.Analysis,
				ExecErr:  runRes.Error,
			})
		}
		return output.PrintJSON(output.JSONParams{
			Result:    runRes.Result,
			ExpectRes: runRes.ExpectRes,
			Analysis:  runRes.Analysis,
			ExecErr:   runRes.Error,
		})
	}

	if err := app.printResult(runRes.Result, runRes.ExpectRes, ctx.path, printResultOptions{skipMeta: ctx.verbose}); err != nil {
		return err
	}

	if runRes.Error != nil {
		if ctx.strict || ctx.returnErrors {
			return runRes.Error
		}
		fmt.Println(color.Red(runRes.Error.Error()))
		return nil
	}

	out, noColor = app.io(ctx.strict)
	validation.PrintWarnings(runRes.Analysis, out, noColor)
	return nil
}

// executeChain handles chain config execution and output.
func (app *rootCommand) executeChain(ctx runContext, runRes *core.RunConfigResult, opts runner.Options) error {
	chainResult, chainErr := app.engine.RunChain(context.Background(), runRes.Analysis.Base, runRes.Analysis.Chain, opts, runRes.Analysis)

	// Handle JSON output mode for chains
	if ctx.jsonOutput {
		return output.PrintJSON(output.JSONParams{ChainResult: chainResult, ExecErr: chainErr})
	}

	// Print results from all completed steps (even if chain failed)
	var outputSavedErr error
	if chainResult != nil {
		for i, stepResult := range chainResult.Results {
			fmt.Fprintf(os.Stderr, "\n--- Step %d: %s ---\n", i+1, chainResult.StepNames[i])
			var expectRes *runner.ExpectationResult
			if i < len(chainResult.ExpectationResults) {
				expectRes = chainResult.ExpectationResults[i]
			}
			if err := app.printResult(stepResult, expectRes, ctx.path, printResultOptions{}); err != nil {
				outputSavedErr = err
			}
		}
	}

	if chainErr != nil {
		if ctx.strict || ctx.returnErrors {
			return chainErr
		}
		fmt.Println(color.Red(chainErr.Error()))
		return nil
	}

	fmt.Fprintln(os.Stderr, "\nChain completed successfully.")
	out, noColor := app.io(ctx.strict)
	validation.PrintWarnings(runRes.Analysis, out, noColor)
	return outputSavedErr
}

// runConfigPathE runs a config file in strict mode (returns error)
func (app *rootCommand) runConfigPathE(path string) error {
	return app.executeRunE(runContext{path: path, strict: true})
}

// runConfigPathWithEnvAndJSONE runs a config file with a specific environment and optional JSON output in strict mode
//
//nolint:unused
func (app *rootCommand) runConfigPathWithEnvAndJSONE(path string, envName string, jsonOutput bool) error {
	return app.executeRunE(runContext{path: path, strict: true, envName: envName, jsonOutput: jsonOutput})
}

// runConfigPathWithOptionsE runs a config file with all options
func (app *rootCommand) runConfigPathWithOptionsE(path string, envName string, jsonOutput bool, strictEnv bool, verbose bool) error {
	return app.executeRunE(runContext{path: path, strict: true, envName: envName, jsonOutput: jsonOutput, strictEnv: strictEnv, verbose: verbose})
}

// printExpectationResult prints expectation results to stderr
func printExpectationResult(res *runner.ExpectationResult) {
	if res.AssertionsTotal == 0 && !res.StatusChecked {
		return
	}

	fmt.Fprintln(os.Stderr)

	// Status check result
	if res.StatusChecked {
		if res.StatusPassed {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Green("[PASS]"), "status check")
		} else {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Red("[FAIL]"), "status check")
		}
	}

	// Print each assertion result
	for _, ar := range res.AssertionResults {
		if ar.Passed {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Green("[PASS]"), ar.Expression)
		} else {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Red("[FAIL]"), ar.Expression)
		}
	}

	// Summary line
	if res.AssertionsTotal > 0 {
		summary := fmt.Sprintf("assertions: %d/%d passed", res.AssertionsPassed, res.AssertionsTotal)
		if res.AllPassed() {
			fmt.Fprintf(os.Stderr, "\n%s\n", color.Green(summary))
		} else {
			fmt.Fprintf(os.Stderr, "\n%s\n", color.Red(summary))
		}
	}
}

// printResultMeta prints request URL and timing to stderr
func printResultMeta(result *runner.Result) {
	if result.RequestURL != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", color.Dim("URL: "+result.RequestURL))
	}
	fmt.Fprintf(os.Stderr, "%s\n", color.Dim("Time: "+result.Duration.String()))
	fmt.Fprintf(os.Stderr, "%s\n", color.Dim(fmt.Sprintf("Size: %s (%d lines, %d chars)", formatBytes(result.BodyBytes), result.BodyLines, result.BodyChars)))
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/langserver"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/validation"
)

// Set via ldflags at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func init() {
	if version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				commit = s.Value[:7]
			}
		case "vcs.time":
			date = s.Value
		}
	}
}

type rootCommand struct {
	urlOverride string
	noColor     bool
	httpClient  *http.Client
	engine      *core.Engine
}

func main() {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	app := &rootCommand{
		httpClient: httpClient,
		engine:     core.NewEngine(httpClient),
	}

	rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			color.SetNoColor(app.noColor)
		},
		Run: app.runInteractive,
	}

	rootCmd.PersistentFlags().StringVarP(&app.urlOverride, "url", "u", "", "Override the URL specified in the config file")
	rootCmd.PersistentFlags().BoolVar(&app.noColor, "no-color", false, "Disable color output")

	rootCmd.AddCommand(app.newRunCmd())
	rootCmd.AddCommand(app.newWatchCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newLSPCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newValidateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func (app *rootCommand) runInteractive(cmd *cobra.Command, args []string) {
	selectedPath, err := tui.FindConfigFileSingle()
	if err != nil {
		log.Fatalf("Failed to select config file: %v", err)
	}
	app.runConfigPath(selectedPath)
}

func (app *rootCommand) newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <file>",
		Short: "Run a request defined in a yapi config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			app.runConfigPath(args[0])
		},
	}
	return cmd
}

func (app *rootCommand) newWatchCmd() *cobra.Command {
	var pretty bool
	var noPretty bool

	cmd := &cobra.Command{
		Use:   "watch [file]",
		Short: "Watch a yapi config file and re-run on changes",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var path string
			interactive := len(args) == 0

			if interactive {
				selectedPath, err := tui.FindConfigFileSingle()
				if err != nil {
					log.Fatalf("Failed to select config file: %v", err)
				}
				path = selectedPath
			} else {
				path = args[0]
			}

			usePretty := pretty || (interactive && !noPretty)

			if usePretty {
				if err := tui.RunWatch(path); err != nil {
					log.Fatalf("Watch failed: %v", err)
				}
			} else {
				app.watchConfigPath(path)
			}
		},
	}

	cmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Enable pretty TUI mode")
	cmd.Flags().BoolVar(&noPretty, "no-pretty", false, "Disable pretty TUI mode")

	return cmd
}

func (app *rootCommand) watchConfigPath(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to resolve path: %v", err)
	}

	clearScreen()
	printWatchHeader(absPath)
	app.runConfigPathSafe(absPath)

	lastMod := getModTime(absPath)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		currentMod := getModTime(absPath)
		if currentMod != lastMod {
			lastMod = currentMod
			clearScreen()
			printWatchHeader(absPath)
			app.runConfigPathSafe(absPath)
		}
	}
}

func getModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printWatchHeader(path string) {
	fmt.Printf("%s\n\n", color.Accent("yapi watch"))
	fmt.Printf("%s\n", color.Dim("[watching "+filepath.Base(path)+"]"))
	fmt.Printf("%s\n\n", color.Dim("["+time.Now().Format("15:04:05")+"]"))
}

// runContext holds options for executeRun
type runContext struct {
	path   string
	strict bool // If true, os.Exit(1) on errors; if false, print and return
}

// executeRun is the unified execution pipeline for both Run and Watch modes.
func (app *rootCommand) executeRun(ctx runContext) {
	opts := runner.Options{
		URLOverride: app.urlOverride,
		NoColor:     app.noColor,
	}

	runRes := app.engine.RunConfig(context.Background(), ctx.path, opts)

	// Handle validation/parse errors first
	if runRes.Error != nil && runRes.Analysis == nil {
		app.handleError(runRes.Error, ctx.strict)
		return
	}

	app.printErrors(runRes.Analysis, ctx.strict)
	if runRes.Analysis != nil && runRes.Analysis.HasErrors() {
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	// Check if this is a chain config
	if runRes.Analysis != nil && len(runRes.Analysis.Chain) > 0 {
		chainResult, chainErr := app.engine.RunChain(context.Background(), runRes.Analysis.Base, runRes.Analysis.Chain, opts)

		// Print results from all completed steps (even if chain failed)
		if chainResult != nil {
			for i, stepResult := range chainResult.Results {
				fmt.Fprintf(os.Stderr, "\n--- Step %d: %s ---\n", i+1, chainResult.StepNames[i])
				// Trim trailing whitespace to avoid double newlines (e.g. TCP responses with \n)
				body := strings.TrimRight(output.Highlight(stepResult.Body, stepResult.ContentType, app.noColor), "\n\r")
				fmt.Println(body)
				printResultMeta(stepResult)
				// Print expectation results for this step
				if i < len(chainResult.ExpectationResults) {
					printExpectationResult(chainResult.ExpectationResults[i])
				}
			}
		}

		if chainErr != nil {
			app.handleError(chainErr, ctx.strict)
			return
		}

		if ctx.strict {
			logHistory(ctx.path, app.urlOverride)
		}

		fmt.Fprintln(os.Stderr, "\nChain completed successfully.")
		app.printWarnings(runRes.Analysis, ctx.strict)
		return
	}

	if runRes.Analysis == nil || runRes.Analysis.Request == nil {
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	if ctx.strict {
		logHistory(ctx.path, app.urlOverride)
	}

	if runRes.Result != nil {
		// Trim trailing whitespace to avoid double newlines
		body := strings.TrimRight(output.Highlight(runRes.Result.Body, runRes.Result.ContentType, app.noColor), "\n\r")
		fmt.Println(body)
		printResultMeta(runRes.Result)
	}

	// Print expectation results
	if runRes.ExpectRes != nil {
		printExpectationResult(runRes.ExpectRes)
	}

	// Handle expectation errors after printing result
	if runRes.Error != nil {
		app.handleError(runRes.Error, ctx.strict)
		return
	}

	app.printWarnings(runRes.Analysis, ctx.strict)
}

// handleError prints an error, optionally exiting for strict mode
func (app *rootCommand) handleError(err error, strict bool) {
	if strict {
		log.Fatalf("%v", err)
	} else {
		fmt.Println(color.Red(err.Error()))
	}
}

// formatDiagnostic formats a single diagnostic with color.
func formatDiagnostic(d validation.Diagnostic) string {
	lineInfo := ""
	if d.Line >= 0 {
		lineInfo = fmt.Sprintf(" (line %d)", d.Line+1)
	}

	switch d.Severity {
	case validation.SeverityError:
		return color.Red("[ERROR]" + lineInfo + " " + d.Message)
	case validation.SeverityWarning:
		return color.Yellow("[WARN]" + lineInfo + " " + d.Message)
	default:
		return color.Cyan("[INFO]" + lineInfo + " " + d.Message)
	}
}

// printDiagnostics prints diagnostics filtered by a predicate.
func (app *rootCommand) printDiagnostics(
	analysis *validation.Analysis,
	strict bool,
	filter func(validation.Diagnostic) bool,
) {
	if analysis == nil {
		return
	}

	out := os.Stdout
	if strict {
		out = os.Stderr
	}

	for _, d := range analysis.Diagnostics {
		if !filter(d) {
			continue
		}
		fmt.Fprintln(out, formatDiagnostic(d))
	}
}

func (app *rootCommand) printErrors(a *validation.Analysis, strict bool) {
	app.printDiagnostics(a, strict, func(d validation.Diagnostic) bool {
		return d.Severity == validation.SeverityError
	})
}

func (app *rootCommand) printWarnings(a *validation.Analysis, strict bool) {
	if a == nil {
		return
	}

	out := os.Stdout
	if strict {
		out = os.Stderr
	}

	for _, w := range a.Warnings {
		fmt.Fprintln(out, color.Yellow("[WARN] "+w))
	}

	app.printDiagnostics(a, strict, func(d validation.Diagnostic) bool {
		return d.Severity != validation.SeverityError
	})
}

// runConfigPathSafe runs a config file without exiting on error (for watch mode)
func (app *rootCommand) runConfigPathSafe(path string) {
	app.executeRun(runContext{path: path, strict: false})
}

func newLSPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Run the yapi language server over stdio",
		Run: func(cmd *cobra.Command, args []string) {
			langserver.Run()
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("yapi %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		},
	}
}

func newValidateCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a yapi config file",
		Long:  "Validate a yapi config file and report diagnostics. Use - to read from stdin.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var text string

			if len(args) == 0 || args[0] == "-" {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					if jsonOutput {
						outputValidateError(err)
					} else {
						log.Fatalf("Failed to read stdin: %v", err)
					}
					return
				}
				text = string(data)
			} else {
				data, err := os.ReadFile(args[0])
				if err != nil {
					if jsonOutput {
						outputValidateError(err)
					} else {
						log.Fatalf("Failed to read file: %v", err)
					}
					return
				}
				text = string(data)
			}

			analysis, err := validation.AnalyzeConfigString(text)
			if err != nil {
				if jsonOutput {
					outputValidateError(err)
				} else {
					log.Fatalf("Validation failed: %v", err)
				}
				return
			}

			if jsonOutput {
				json.NewEncoder(os.Stdout).Encode(analysis.ToJSON())
			} else {
				outputValidateText(analysis)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output diagnostics as JSON")

	return cmd
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
	json.NewEncoder(os.Stdout).Encode(out)
}

func outputValidateText(analysis *validation.Analysis) {
	hasOutput := len(analysis.Warnings) > 0 || len(analysis.Diagnostics) > 0

	for _, w := range analysis.Warnings {
		fmt.Println(color.Yellow("[WARN] " + w))
	}

	for _, d := range analysis.Diagnostics {
		fmt.Println(formatDiagnostic(d))
	}

	if !hasOutput {
		fmt.Println(color.Green("Valid"))
	}

	if analysis.HasErrors() {
		os.Exit(1)
	}
}

// runConfigPath runs a config file in strict mode (exits on error)
func (app *rootCommand) runConfigPath(path string) {
	app.executeRun(runContext{path: path, strict: true})
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

func formatBytes(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [count]",
		Short: "Show yapi command history (default: last 10)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			count := 10
			if len(args) == 1 {
				n, err := fmt.Sscanf(args[0], "%d", &count)
				if err != nil || n != 1 || count < 1 {
					log.Fatalf("Invalid count: %s", args[0])
				}
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("Failed to get home directory: %v", err)
			}

			historyFile := filepath.Join(homeDir, ".yapi_history")
			data, err := os.ReadFile(historyFile)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No history yet")
					return
				}
				log.Fatalf("Failed to read history: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
				fmt.Println("No history yet")
				return
			}

			start := len(lines) - count
			if start < 0 {
				start = 0
			}

			for _, line := range lines[start:] {
				fmt.Println(line)
			}
		},
	}
	return cmd
}

// logHistory writes the executed command to ~/.yapi_history for shell integration
func logHistory(configPath, urlOverride string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyFile := filepath.Join(homeDir, ".yapi_history")
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		absPath = configPath
	}

	cmd := fmt.Sprintf("yapi run \"%s\"", absPath)
	if urlOverride != "" {
		cmd += fmt.Sprintf(" -u \"%s\"", urlOverride)
	}

	line := fmt.Sprintf("%d | %s\n", time.Now().Unix(), cmd)
	f.WriteString(line)
}

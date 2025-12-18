package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/cli/commands"
	"yapi.run/cli/internal/cli/middleware"
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/langserver"
	"yapi.run/cli/internal/observability"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/share"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/utils"
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
	urlOverride  string
	noColor      bool
	binaryOutput bool
	httpClient   *http.Client
	engine       *core.Engine
}

// io returns the appropriate writer and color flag based on strict mode
func (app *rootCommand) io(strict bool) (io.Writer, bool) {
	if strict {
		return os.Stderr, app.noColor
	}
	return os.Stdout, app.noColor
}

// selectConfigFile returns the config file path, handling interactive TUI selection when no args provided.
// Returns (selectedPath, fromTUI, error).
func selectConfigFile(args []string, cmdName string) (string, bool, error) {
	if len(args) > 0 {
		return args[0], false, nil
	}

	selectedPath, err := tui.FindConfigFileSingle()
	if err != nil {
		return "", false, fmt.Errorf("failed to select config file: %w", err)
	}

	// Log to history with from_tui flag
	absPath, _ := filepath.Abs(selectedPath)
	logHistoryFromTUI(fmt.Sprintf("yapi %s %q", cmdName, absPath))

	return selectedPath, true, nil
}

func main() {
	observability.Init(version, commit)
	defer observability.Close()

	// Wire observability hook - main.go is the composition root
	requestHook := func(stats map[string]any) {
		observability.Track("request_executed", stats)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	app := &rootCommand{
		httpClient: httpClient,
		engine:     core.NewEngine(httpClient, core.WithRequestHook(requestHook)),
	}

	cfg := &commands.Config{}
	handlers := &commands.Handlers{
		RunInteractive: app.runInteractiveE,
		Run:            app.runE,
		Watch:          app.watchE,
		History:        historyE,
		LSP:            lspE,
		Version:        versionE,
		Validate:       validateE,
		Share:          shareE,
		Test:           app.testE,
	}

	rootCmd := commands.BuildRoot(cfg, handlers)

	// Wire up the config to app after flags are parsed
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		app.urlOverride = cfg.URLOverride
		app.noColor = cfg.NoColor
		app.binaryOutput = cfg.BinaryOutput
		color.SetNoColor(app.noColor)
	}
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		// Log command to history (skip meta commands)
		switch cmd.Name() {
		case "history", "version", "lsp", "help", "yapi":
			return
		}
		logHistoryCmd(reconstructCommand(cmd, args))
	}

	// Wrap all commands with observability middleware
	middleware.WrapWithObservability(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.Red(err.Error()))
		os.Exit(1)
	}
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
	return app.runConfigPathE(path)
}

func (app *rootCommand) watchE(cmd *cobra.Command, args []string) error {
	pretty, _ := cmd.Flags().GetBool("pretty")
	noPretty, _ := cmd.Flags().GetBool("no-pretty")

	path, fromTUI, err := selectConfigFile(args, "watch")
	if err != nil {
		return err
	}

	usePretty := pretty || (fromTUI && !noPretty)

	if usePretty {
		return tui.RunWatch(path)
	}
	return app.watchConfigPath(path)
}

func (app *rootCommand) watchConfigPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	clearScreen()
	printWatchHeader(absPath)
	app.runConfigPathSafe(absPath)

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
			app.runConfigPathSafe(absPath)
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

// runContext holds options for executeRun
type runContext struct {
	path   string
	strict bool // If true, return error on failures; if false, print and return nil
}

// printResult outputs a single result with optional expectation.
func (app *rootCommand) printResult(result *runner.Result, expectRes *runner.ExpectationResult) {
	if result != nil {
		// Check if stdout is a TTY (terminal)
		isTTY := isTerminal(os.Stdout)

		// Check if content is binary
		isBinary := utils.IsBinaryContent(result.Body)

		// Skip dumping binary to terminal unless explicitly requested or piping
		if isBinary && isTTY && !app.binaryOutput {
			fmt.Fprintf(os.Stderr, "\n%s\n", color.Yellow("Binary content detected. Output hidden to prevent terminal corruption."))
			fmt.Fprintf(os.Stderr, "%s\n", color.Dim("To display binary output, use --binary-output flag or pipe to a file."))
		} else {
			body := strings.TrimRight(output.Highlight(result.Body, result.ContentType, app.noColor), "\n\r")
			fmt.Println(body)
		}

		printResultMeta(result)
	}
	if expectRes != nil {
		printExpectationResult(expectRes)
	}
}

// executeRunE is the unified execution pipeline for both Run and Watch modes.
// Returns error for middleware to capture.
func (app *rootCommand) executeRunE(ctx runContext) error {
	opts := runner.Options{
		URLOverride:  app.urlOverride,
		NoColor:      app.noColor,
		BinaryOutput: app.binaryOutput,
	}

	runRes := app.engine.RunConfig(context.Background(), ctx.path, opts)

	// Handle validation/parse errors first
	if runRes.Error != nil && runRes.Analysis == nil {
		if ctx.strict {
			return runRes.Error
		}
		fmt.Println(color.Red(runRes.Error.Error()))
		return nil
	}

	out, noColor := app.io(ctx.strict)
	validation.PrintErrors(runRes.Analysis, out, noColor)
	if runRes.Analysis != nil && runRes.Analysis.HasErrors() {
		if ctx.strict {
			return &validation.Error{Diagnostics: runRes.Analysis.Diagnostics}
		}
		return nil
	}

	// Check if this is a chain config
	if runRes.Analysis != nil && len(runRes.Analysis.Chain) > 0 {
		chainResult, chainErr := app.engine.RunChain(context.Background(), runRes.Analysis.Base, runRes.Analysis.Chain, opts, runRes.Analysis)

		// Print results from all completed steps (even if chain failed)
		if chainResult != nil {
			for i, stepResult := range chainResult.Results {
				fmt.Fprintf(os.Stderr, "\n--- Step %d: %s ---\n", i+1, chainResult.StepNames[i])
				var expectRes *runner.ExpectationResult
				if i < len(chainResult.ExpectationResults) {
					expectRes = chainResult.ExpectationResults[i]
				}
				app.printResult(stepResult, expectRes)
			}
		}

		if chainErr != nil {
			if ctx.strict {
				return chainErr
			}
			fmt.Println(color.Red(chainErr.Error()))
			return nil
		}

		fmt.Fprintln(os.Stderr, "\nChain completed successfully.")
		out, noColor := app.io(ctx.strict)
		validation.PrintWarnings(runRes.Analysis, out, noColor)
		return nil
	}

	if runRes.Analysis == nil || runRes.Analysis.Request == nil {
		if ctx.strict {
			return errors.New("invalid config")
		}
		return nil
	}

	app.printResult(runRes.Result, runRes.ExpectRes)

	if runRes.Error != nil {
		if ctx.strict {
			return runRes.Error
		}
		fmt.Println(color.Red(runRes.Error.Error()))
		return nil
	}

	out, noColor = app.io(ctx.strict)
	validation.PrintWarnings(runRes.Analysis, out, noColor)
	return nil
}

// runConfigPathSafe runs a config file without returning error (for watch mode)
func (app *rootCommand) runConfigPathSafe(path string) {
	_ = app.executeRunE(runContext{path: path, strict: false})
}

// runConfigPathE runs a config file in strict mode (returns error)
func (app *rootCommand) runConfigPathE(path string) error {
	return app.executeRunE(runContext{path: path, strict: true})
}

func lspE(cmd *cobra.Command, args []string) error {
	langserver.Run()
	return nil
}

func versionE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		info := map[string]any{
			"version": version,
			"commit":  commit,
			"date":    date,
		}
		return json.NewEncoder(os.Stdout).Encode(info)
	}

	fmt.Printf("yapi %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built:  %s\n", date)
	return nil
}

func validateE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	path, _, err := selectConfigFile(args, "validate")
	if err != nil {
		if jsonOutput {
			outputValidateError(err)
			return nil
		}
		return err
	}

	data, err := utils.ReadInput(path)
	if err != nil {
		if jsonOutput {
			outputValidateError(err)
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	analysis, err := validation.AnalyzeConfigString(string(data))
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
	analysis, analysisErr := validation.AnalyzeConfigString(content)
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

type historyEntry struct {
	Timestamp string   `json:"timestamp"`
	Event     string   `json:"event,omitempty"`  // legacy single event
	Events    []string `json:"events,omitempty"` // new merged events
	Command   string   `json:"command,omitempty"`
	FromTUI   bool     `json:"from_tui,omitempty"`
	// Fields from request tracking
	OS      string         `json:"os,omitempty"`
	Arch    string         `json:"arch,omitempty"`
	Version string         `json:"version,omitempty"`
	Commit  string         `json:"commit,omitempty"`
	Props   map[string]any `json:"-"` // For parsing additional fields
}

func historyE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	count := 10
	if len(args) == 1 {
		n, err := fmt.Sscanf(args[0], "%d", &count)
		if err != nil || n != 1 || count < 1 {
			return fmt.Errorf("invalid count: %s", args[0])
		}
	}

	data, err := os.ReadFile(observability.HistoryFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No history yet")
			return nil
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("No history yet")
		return nil
	}

	start := len(lines) - count
	if start < 0 {
		start = 0
	}

	entries := lines[start:]

	if jsonOutput {
		fmt.Println("[")
		for i, line := range entries {
			fmt.Print("  " + line)
			if i < len(entries)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
		return nil
	}

	// Pretty print for humans
	for _, line := range entries {
		var entry historyEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		t, _ := time.Parse(time.RFC3339, entry.Timestamp)
		timeStr := color.Dim(t.Format("2006-01-02 15:04:05"))

		// New merged format has Command field directly
		if entry.Command != "" {
			fmt.Printf("%s  %s\n", timeStr, entry.Command)
			continue
		}

		// Legacy: request_executed entries had method/url
		if entry.Event == "request_executed" {
			var raw map[string]any
			if err := json.Unmarshal([]byte(line), &raw); err == nil {
				method, _ := raw["method"].(string)
				url, _ := raw["url"].(string)
				status, _ := raw["status_code"].(float64)
				if method != "" && url != "" {
					fmt.Printf("%s  %s %s %s\n", timeStr, color.Cyan(method), url, color.Dim(fmt.Sprintf("[%d]", int(status))))
					continue
				}
			}
		}
	}
	return nil
}

// logHistoryCmd writes a command to history as JSON
func logHistoryCmd(cmdStr string) {
	logHistoryEntry(cmdStr, false)
}

// logHistoryFromTUI writes a TUI-selected command to history
func logHistoryFromTUI(cmdStr string) {
	logHistoryEntry(cmdStr, true)
}

func logHistoryEntry(cmdStr string, fromTUI bool) {
	props := map[string]any{
		"command": cmdStr,
	}
	if fromTUI {
		props["from_tui"] = true
	}
	observability.Track("command", props)
}

// reconstructCommand builds the full command string from cobra command and args
func reconstructCommand(cmd *cobra.Command, args []string) string {
	parts := []string{"yapi", cmd.Name()}

	// Add flags that were set
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			parts = append(parts, "--"+f.Name)
		} else {
			parts = append(parts, fmt.Sprintf("--%s=%q", f.Name, f.Value.String()))
		}
	})

	// Add args (quote paths)
	for _, arg := range args {
		absPath, err := filepath.Abs(arg)
		if err == nil && fileExists(absPath) {
			parts = append(parts, fmt.Sprintf("%q", absPath))
		} else {
			parts = append(parts, arg)
		}
	}

	return strings.Join(parts, " ")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (app *rootCommand) testE(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Determine search directory
	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	// Find all test files
	testFiles, err := findTestFiles(searchDir)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	if len(testFiles) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.test.yapi.yml files found"))
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s\n\n", color.Accent(fmt.Sprintf("Running %d test(s)...", len(testFiles))))

	// Run each test and collect results
	type testResult struct {
		file   string
		passed bool
		err    error
	}

	var results []testResult
	passCount := 0

	for i, testFile := range testFiles {
		relPath, _ := filepath.Rel(searchDir, testFile)
		if verbose {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Dim(fmt.Sprintf("[%d/%d]", i+1, len(testFiles))), relPath)
		}

		// Run the test file
		err := app.executeRunE(runContext{path: testFile, strict: true})

		result := testResult{
			file:   relPath,
			passed: err == nil,
			err:    err,
		}
		results = append(results, result)

		if err == nil {
			passCount++
			if !verbose {
				fmt.Fprintf(os.Stderr, "%s ", color.Green("✓"))
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n\n", color.Green("PASS"))
			}
		} else {
			if !verbose {
				fmt.Fprintf(os.Stderr, "%s ", color.Red("✗"))
			} else {
				fmt.Fprintf(os.Stderr, "  %s %s\n\n", color.Red("FAIL"), color.Dim(err.Error()))
			}
		}
	}

	if !verbose {
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\n")
	if passCount == len(results) {
		fmt.Fprintf(os.Stderr, "%s\n", color.Green(fmt.Sprintf("All %d test(s) passed", passCount)))
		return nil
	}

	failCount := len(results) - passCount
	fmt.Fprintf(os.Stderr, "%s\n", color.Red(fmt.Sprintf("%d of %d test(s) failed", failCount, len(results))))

	// List failed tests
	fmt.Fprintf(os.Stderr, "\n%s\n", color.Red("Failed tests:"))
	for _, r := range results {
		if !r.passed {
			fmt.Fprintf(os.Stderr, "  %s %s\n", color.Red("✗"), r.file)
			if r.err != nil && verbose {
				fmt.Fprintf(os.Stderr, "    %s\n", color.Dim(r.err.Error()))
			}
		}
	}

	return fmt.Errorf("%d test(s) failed", failCount)
}

// findTestFiles recursively finds all *.test.yapi.yml files in the given directory
func findTestFiles(dir string) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml" {
			base := filepath.Base(path)
			// Match *.test.yapi.yml or *.test.yapi.yaml
			if strings.HasSuffix(base, ".test.yapi.yml") || strings.HasSuffix(base, ".test.yapi.yaml") {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})

	return testFiles, err
}

// isTerminal checks if the given file is a terminal (TTY)
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

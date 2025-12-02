package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/langserver"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/validation"
)

type rootCommand struct {
	urlOverride     string
	noColor         bool
	httpClient      *http.Client
	executorFactory *executor.Factory
}

func main() {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	app := &rootCommand{
		httpClient:      httpClient,
		executorFactory: executor.NewFactory(httpClient),
	}
	rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
		Run:   app.runInteractive,
	}

	rootCmd.PersistentFlags().StringVarP(&app.urlOverride, "url", "u", "", "Override the URL specified in the config file")
	rootCmd.PersistentFlags().BoolVar(&app.noColor, "no-color", false, "Disable color output")

	rootCmd.AddCommand(app.newRunCmd())
	rootCmd.AddCommand(app.newWatchCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newLSPCmd())

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

	// Initial run
	clearScreen()
	printWatchHeader(absPath)
	app.runConfigPathSafe(absPath)

	// Get initial mod time
	lastMod := getModTime(absPath)

	// Poll for changes
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
	fmt.Printf("\033[2m[watching %s]\033[0m\n", filepath.Base(path))
	fmt.Printf("\033[2m[%s]\033[0m\n\n", time.Now().Format("15:04:05"))
}

// runContext holds options for executeRun
type runContext struct {
	path   string
	strict bool // If true, os.Exit(1) on errors; if false, print and return
}

// executeRun is the unified execution pipeline for both Run and Watch modes.
func (app *rootCommand) executeRun(ctx runContext) {
	analysis, err := validation.AnalyzeConfigFile(ctx.path)
	if err != nil {
		app.handleError(err, ctx.strict)
		return
	}

	// Print errors only (warnings come after output)
	app.printErrors(analysis, ctx.strict)

	if analysis.HasErrors() {
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	req := analysis.Request
	if req == nil {
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	if ctx.strict {
		logHistory(ctx.path, app.urlOverride)
	}

	exec, err := app.createExecutor(req.Metadata["transport"])
	if err != nil {
		app.handleError(err, ctx.strict)
		return
	}

	opts := runner.Options{
		URLOverride: app.urlOverride,
		NoColor:     app.noColor,
	}

	output, result, err := runner.RunAndFormat(context.Background(), exec, req, nil, opts)
	if err != nil {
		app.handleError(err, ctx.strict)
		return
	}

	fmt.Println(output)
	printResultMeta(result)
	app.printWarnings(analysis, ctx.strict)
}

// handleError prints an error, optionally exiting for strict mode
func (app *rootCommand) handleError(err error, strict bool) {
	if strict {
		log.Fatalf("%v", err)
	} else {
		fmt.Printf("\033[31m%v\033[0m\n", err)
	}
}

// printErrors prints only error-level diagnostics (before request runs)
func (app *rootCommand) printErrors(analysis *validation.Analysis, strict bool) {
	out := os.Stdout
	if strict {
		out = os.Stderr
	}

	for _, d := range analysis.Diagnostics {
		if d.Severity != validation.SeverityError {
			continue
		}
		lineInfo := ""
		if d.Line >= 0 {
			lineInfo = fmt.Sprintf(" (line %d)", d.Line+1)
		}
		fmt.Fprintf(out, "\033[31m[ERROR]%s %s\033[0m\n", lineInfo, d.Message)
	}
}

// printWarnings prints warnings and info diagnostics (after output)
func (app *rootCommand) printWarnings(analysis *validation.Analysis, strict bool) {
	out := os.Stdout
	if strict {
		out = os.Stderr
	}

	for _, w := range analysis.Warnings {
		fmt.Fprintf(out, "\033[33m[WARN] %s\033[0m\n", w)
	}

	for _, d := range analysis.Diagnostics {
		if d.Severity == validation.SeverityError {
			continue
		}
		prefix := "[INFO]"
		color := "\033[36m"
		if d.Severity == validation.SeverityWarning {
			prefix = "[WARN]"
			color = "\033[33m"
		}
		lineInfo := ""
		if d.Line >= 0 {
			lineInfo = fmt.Sprintf(" (line %d)", d.Line+1)
		}
		fmt.Fprintf(out, "%s%s%s %s\033[0m\n", color, prefix, lineInfo, d.Message)
	}
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

// runConfigPath runs a config file in strict mode (exits on error)
func (app *rootCommand) runConfigPath(path string) {
	app.executeRun(runContext{path: path, strict: true})
}

func (app *rootCommand) createExecutor(transport string) (executor.Executor, error) {
	return app.executorFactory.Create(transport)
}

// dim wraps text in ANSI dim escape codes
func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

// printResultMeta prints request URL and timing to stderr
func printResultMeta(result *runner.Result) {
	if result.RequestURL != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", dim("URL: "+result.RequestURL))
	}
	fmt.Fprintf(os.Stderr, "%s\n", dim("Time: "+result.Duration.String()))
	fmt.Fprintf(os.Stderr, "%s\n", dim(fmt.Sprintf("Size: %s (%d lines, %d chars)", formatBytes(result.BodyBytes), result.BodyLines, result.BodyChars)))
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

			// Get last N lines
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

	// Get absolute path for the config
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		absPath = configPath
	}

	// Build the command string
	cmd := fmt.Sprintf("yapi run \"%s\"", absPath)
	if urlOverride != "" {
		cmd += fmt.Sprintf(" -u \"%s\"", urlOverride)
	}

	// Write in format: <timestamp> | <command>
	line := fmt.Sprintf("%d | %s\n", time.Now().Unix(), cmd)
	f.WriteString(line)
}

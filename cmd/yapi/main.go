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
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/langserver"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/validation"
)

// ANSI color codes (matching theme orange accent #ff9e64)
const (
	colorOrange = "\033[38;2;255;158;100m"
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
)

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
	fmt.Printf("%s🐑 yapi watch%s\n\n", colorOrange, colorReset)
	fmt.Printf("%s[watching %s]%s\n", colorDim, filepath.Base(path), colorReset)
	fmt.Printf("%s[%s]%s\n\n", colorDim, time.Now().Format("15:04:05"), colorReset)
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

	analysis, result, err := app.engine.RunConfig(context.Background(), ctx.path, opts)
	if err != nil {
		app.handleError(err, ctx.strict)
		return
	}

	if analysis == nil || analysis.Request == nil {
		app.printErrors(analysis, ctx.strict)
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	app.printErrors(analysis, ctx.strict)
	if analysis.HasErrors() {
		if ctx.strict {
			os.Exit(1)
		}
		return
	}

	if ctx.strict {
		logHistory(ctx.path, app.urlOverride)
	}

	if result != nil {
		fmt.Println(output.Highlight(result.Body, result.ContentType, app.noColor))
		printResultMeta(result)
	}
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
		color, prefix := "\033[36m", "[INFO]"
		if d.Severity == validation.SeverityWarning {
			color, prefix = "\033[33m", "[WARN]"
		}
		if d.Severity == validation.SeverityError {
			color, prefix = "\033[31m", "[ERROR]"
		}

		lineInfo := ""
		if d.Line >= 0 {
			lineInfo = fmt.Sprintf(" (line %d)", d.Line+1)
		}
		fmt.Fprintf(out, "%s%s%s %s\033[0m\n", color, prefix, lineInfo, d.Message)
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
		fmt.Fprintf(out, "\033[33m[WARN] %s\033[0m\n", w)
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

// runConfigPath runs a config file in strict mode (exits on error)
func (app *rootCommand) runConfigPath(path string) {
	app.executeRun(runContext{path: path, strict: true})
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

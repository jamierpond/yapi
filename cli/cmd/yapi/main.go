package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cli/internal/config"
	"cli/internal/executor"
	"cli/internal/filter"
	"cli/internal/langserver"
	"cli/internal/output"
	"cli/internal/tui"
	"cli/internal/validation"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	urlOverride string
	noColor     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
		Run:   runInteractive,
	}

	rootCmd.PersistentFlags().StringVarP(&urlOverride, "url", "u", "", "Override the URL specified in the config file")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newLSPCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runInteractive(cmd *cobra.Command, args []string) {
	selectedPath, err := tui.FindConfigFileSingle()
	if err != nil {
		log.Fatalf("Failed to select config file: %v", err)
	}
	runConfigPath(selectedPath)
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <file>",
		Short: "Run a request defined in a yapi config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runConfigPath(args[0])
		},
	}
	return cmd
}

func newWatchCmd() *cobra.Command {
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

			// --pretty forces pretty mode, --no-pretty disables it
			// Default: pretty in interactive mode, simple when file is passed
			usePretty := pretty || (interactive && !noPretty)

			if usePretty {
				if err := tui.RunWatch(path); err != nil {
					log.Fatalf("Watch failed: %v", err)
				}
			} else {
				watchConfigPath(path)
			}
		},
	}

	cmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Enable pretty TUI mode")
	cmd.Flags().BoolVar(&noPretty, "no-pretty", false, "Disable pretty TUI mode")

	return cmd
}

func watchConfigPath(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to resolve path: %v", err)
	}

	// Initial run
	clearScreen()
	printWatchHeader(absPath)
	runConfigPathSafe(absPath)

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
			runConfigPathSafe(absPath)
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

func runConfigPathSafe(path string) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		fmt.Printf("\033[31mError loading config: %v\033[0m\n", err)
		return
	}

	issues := validation.ValidateConfig(cfg)
	hasErrors := false
	for _, issue := range issues {
		lvl := "WARN"
		color := "\033[33m" // yellow
		if issue.Severity == validation.SeverityError {
			lvl = "ERROR"
			color = "\033[31m" // red
			hasErrors = true
		}
		if issue.Field != "" {
			fmt.Printf("%s[%s] %s: %s\033[0m\n", color, lvl, issue.Field, issue.Message)
		} else {
			fmt.Printf("%s[%s] %s\033[0m\n", color, lvl, issue.Message)
		}
	}
	if hasErrors {
		fmt.Println("\033[31mConfig validation failed\033[0m")
		return
	}

	if urlOverride != "" {
		cfg.URL = urlOverride
	}

	body, ctype, err := executeConfig(cfg)
	if err != nil {
		fmt.Printf("\033[31mRequest failed: %v\033[0m\n", err)
		return
	}

	if cfg.JQFilter != "" {
		body, err = filter.ApplyJQ(body, cfg.JQFilter)
		if err != nil {
			fmt.Printf("\033[31mJQ filter failed: %v\033[0m\n", err)
			return
		}
		ctype = "application/json"
	}

	fmt.Println(output.Highlight(body, ctype, noColor))
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

func runConfigPath(path string) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	issues := validation.ValidateConfig(cfg)
	hasErrors := false
	for _, issue := range issues {
		lvl := "WARN"
		if issue.Severity == validation.SeverityError {
			lvl = "ERROR"
		}
		if issue.Field != "" {
			log.Printf("[%s] %s: %s", lvl, issue.Field, issue.Message)
		} else {
			log.Printf("[%s] %s", lvl, issue.Message)
		}
		if issue.Severity == validation.SeverityError {
			hasErrors = true
		}
	}
	if hasErrors {
		log.Fatal("Config validation failed")
	}

	if urlOverride != "" {
		cfg.URL = urlOverride
	}

	logHistory(path, urlOverride)

	body, ctype, err := executeConfig(cfg)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if cfg.JQFilter != "" {
		body, err = filter.ApplyJQ(body, cfg.JQFilter)
		if err != nil {
			log.Fatalf("JQ filter failed: %v", err)
		}
		ctype = "application/json"
	}

	fmt.Println(output.Highlight(body, ctype, noColor))
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

// executeConfig keeps main() clean and testable.
// Returns: body, contentType, error
func executeConfig(cfg *config.YapiConfig) (string, string, error) {
	// Detect transport from URL scheme
	transport := detectTransport(cfg)

	switch transport {
	case "grpc":
		body, err := executor.NewGRPCExecutor().Execute(cfg)
		return body, "application/json", err
	case "tcp":
		body, err := executor.NewTCPExecutor().Execute(cfg)
		return body, "text/plain", err
	case "http":
		if cfg.Method == "" {
			cfg.Method = "GET"
		}
		resp, err := executor.NewHTTPExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	default:
		return "", "", fmt.Errorf("unsupported transport: %s", transport)
	}
}

// detectTransport determines the transport type from URL scheme or method field
func detectTransport(cfg *config.YapiConfig) string {
	urlLower := strings.ToLower(cfg.URL)

	// Check URL scheme first
	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") {
		return "grpc"
	}
	if strings.HasPrefix(urlLower, "tcp://") {
		return "tcp"
	}

	// Fall back to method field (deprecated but still supported)
	switch cfg.Method {
	case "grpc":
		return "grpc"
	case "tcp":
		return "tcp"
	default:
		return "http"
	}
}

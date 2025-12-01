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
	"cli/internal/output"
	"cli/internal/tui"
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
	rootCmd.AddCommand(newHistoryCmd())

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

	cfg, err := config.LoadConfig(selectedPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if urlOverride != "" {
		cfg.URL = urlOverride
	}

	logHistory(selectedPath, urlOverride)

	result, contentType, err := executeConfig(cfg)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if cfg.JQFilter != "" {
		result, err = filter.ApplyJQ(result, cfg.JQFilter)
		if err != nil {
			log.Fatalf("JQ filter failed: %v", err)
		}
		// jq output is always JSON
		contentType = "application/json"
	}

	fmt.Println(output.Highlight(result, contentType, noColor))
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <file>",
		Short: "Run a request defined in a yapi config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configPath = args[0]

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			if urlOverride != "" {
				cfg.URL = urlOverride
			}

			logHistory(configPath, urlOverride)

			result, contentType, err := executeConfig(cfg)
			if err != nil {
				log.Fatalf("Request failed: %v", err)
			}

			if cfg.JQFilter != "" {
				result, err = filter.ApplyJQ(result, cfg.JQFilter)
				if err != nil {
					log.Fatalf("JQ filter failed: %v", err)
				}
				// jq output is always JSON
				contentType = "application/json"
			}

			fmt.Println(output.Highlight(result, contentType, noColor))
		},
	}
	return cmd
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
	switch cfg.Method {
	case "grpc":
		body, err := executor.NewGRPCExecutor().Execute(cfg)
		return body, "application/json", err
	case "tcp":
		body, err := executor.NewTCPExecutor().Execute(cfg)
		return body, "text/plain", err
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		resp, err := executor.NewHTTPExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	case "":
		cfg.Method = "GET"
		resp, err := executor.NewHTTPExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	default:
		return "", "", fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

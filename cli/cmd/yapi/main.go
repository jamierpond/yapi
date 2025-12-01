package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"cli/internal/config"
	"cli/internal/executor"
	"cli/internal/filter"
	"cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	urlOverride string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
	}

	rootCmd.PersistentFlags().StringVarP(&urlOverride, "url", "u", "", "Override the URL specified in the config file")

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newHistoryCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file]",
		Short: "Run a request defined in a yapi config file",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// If a path is provided, use it directly
			if len(args) == 1 {
				configPath = args[0]
			} else {
				// Interactive mode: pop TUI to select a config file
				selectedPath, err := tui.FindConfigFileSingle()
				if err != nil {
					log.Fatalf("Failed to select config file: %v", err)
				}
				configPath = selectedPath
			}

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			if urlOverride != "" {
				cfg.URL = urlOverride
			}

			// Log to history for shell integration
			logHistory(configPath, urlOverride)

			result, err := executeConfig(cfg)
			if err != nil {
				log.Fatalf("Request failed: %v", err)
			}

			// Apply jq filter if specified
			if cfg.JQFilter != "" {
				result, err = filter.ApplyJQ(result, cfg.JQFilter)
				if err != nil {
					log.Fatalf("JQ filter failed: %v", err)
				}
			}

			// Pure response on stdout, no extra text
			fmt.Println(result)
		},
	}
	return cmd
}

func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show yapi command history",
		Run: func(cmd *cobra.Command, args []string) {
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

			fmt.Print(string(data))
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
	cmd := fmt.Sprintf("yapi \"%s\"", absPath)
	if urlOverride != "" {
		cmd += fmt.Sprintf(" -u \"%s\"", urlOverride)
	}

	// Write in format: <timestamp> | <command>
	line := fmt.Sprintf("%d | %s\n", time.Now().Unix(), cmd)
	f.WriteString(line)
}

// executeConfig keeps main() clean and testable.
func executeConfig(cfg *config.YapiConfig) (string, error) {
	switch cfg.Method {
	case "grpc":
		return executor.NewGRPCExecutor().Execute(cfg)
	case "tcp":
		return executor.NewTCPExecutor().Execute(cfg)
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return executor.NewHTTPExecutor().Execute(cfg)
	case "":
		cfg.Method = "GET"
		return executor.NewHTTPExecutor().Execute(cfg)
	default:
		return "", fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

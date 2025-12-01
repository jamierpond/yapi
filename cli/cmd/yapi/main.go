package main

import (
	"fmt"
	"log"
	"os"

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
		Use:   "yapi [file]",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
		Args:  cobra.MaximumNArgs(1),
		Run:   runYapi,
	}

	rootCmd.PersistentFlags().StringVarP(&urlOverride, "url", "u", "", "Override the URL specified in the config file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runYapi(cmd *cobra.Command, args []string) {
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

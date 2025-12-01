package main

import (
	"fmt"
	"log"
	"os"

	"cli/internal/config"
	"cli/internal/executor"
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

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the yapi config file")
	rootCmd.PersistentFlags().StringVarP(&urlOverride, "url", "u", "", "Override the URL specified in the config file")

	// yapi pick  -> just print selected path(s)
	rootCmd.AddCommand(newPickCmd())
	// yapi run   -> actually execute the request
	rootCmd.AddCommand(newRunCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newPickCmd() *cobra.Command {
	var multi bool

	cmd := &cobra.Command{
		Use:   "pick",
		Short: "Interactively pick a .yapi.yml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := tui.FindConfigFileMulti(multi)
			if err != nil {
				return err
			}
			// Print only the chosen path(s), no decoration
			for _, f := range files {
				fmt.Println(f)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Allow picking multiple files")
	return cmd
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a request defined in a yapi config file",
		Run: func(cmd *cobra.Command, args []string) {
			// If no config path is provided, try to find one.
			// The find function will pop a TUI if it's an interactive session.
			if configPath == "" {
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

			// Pure response on stdout, no extra text
			fmt.Println(result)
		},
	}

	return cmd
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

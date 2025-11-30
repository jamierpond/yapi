package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"yapi/internal/config"
	"yapi/internal/executor"
)

var ( // Global flags
	configPath string
	urlOverride string
)

func main() {
	 rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "yapi is a unified API client for HTTP, gRPC, and TCP",
		Long: `yapi is a command-line tool designed to simplify interactions with various API types,
allowing users to send requests to HTTP, gRPC, and TCP endpoints using a unified configuration.`, // TODO: Improve long description
		Run: func(cmd *cobra.Command, args []string) {
			if configPath == "" {
				// TODO: Implement fuzzy finder for .yapi.yml files
				log.Fatal("No config file specified. Fuzzy finder not yet implemented.")
			}

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			if urlOverride != "" {
				cfg.URL = urlOverride
			}

			// For now, assume HTTP as per the brief's focus
			if cfg.Method == "" { // Default to GET if not specified
				cfg.Method = "GET"
			}
			httpExec := executor.NewHTTPExecutor()
			result, err := httpExec.Execute(cfg)
			if err != nil {
				log.Fatalf("Request failed: %v", err)
			}

			fmt.Println(result)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the yapi config file")
	rootCmd.PersistentFlags().StringVarP(&urlOverride, "url", "u", "", "Override the URL specified in the config file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

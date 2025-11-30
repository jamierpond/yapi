package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"cli/internal/config"
	"cli/internal/executor"
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

			var result string
			switch cfg.Method {
			case "grpc":
				grpcExec := executor.NewGRPCExecutor()
				result, err = grpcExec.Execute(cfg)
			case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS": // Add other HTTP methods as needed
				httpExec := executor.NewHTTPExecutor()
				result, err = httpExec.Execute(cfg)
			case "": // Default to GET if method is not specified in config
				cfg.Method = "GET"
				httpExec := executor.NewHTTPExecutor()
				result, err = httpExec.Execute(cfg)
			default:
				log.Fatalf("Unsupported method: %s", cfg.Method)
			}

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

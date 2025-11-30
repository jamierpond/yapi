package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cliURL  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yapi",
		Short: "A YAML API testing tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. If no config file, trigger Fuzzy Finder (internal/tui)
			if cfgFile == "" {
				// TODO: Implement file picker logic
				fmt.Println("No config selected. (Interactive mode pending implementation)")
				return nil
			}

			// 2. Load and Parse Config (internal/config)
			fmt.Printf("Loading config: %s\n", cfgFile)
			// TODO: LoadYAML(cfgFile)

			// 3. Determine Protocol & Execute (internal/executor)
			// TODO: switch config.Method { ... }

			return nil
		},
	}

	// Flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to YAML config file")
	rootCmd.PersistentFlags().StringVarP(&cliURL, "url", "u", "", "Override base URL")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

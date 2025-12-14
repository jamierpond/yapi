//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"
	"yapi.run/cli/internal/cli/commands"
)

func main() {
	outputDir := "./docs/cli"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	// Build command tree without handlers (for doc generation only)
	rootCmd := commands.BuildRoot(nil, nil)

	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		log.Fatalf("failed to generate docs: %v", err)
	}

	log.Printf("Generated %d docs in %s", len(rootCmd.Commands())+1, outputDir)
}

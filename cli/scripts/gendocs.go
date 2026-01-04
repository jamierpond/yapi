//go:build ignore

package main

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra/doc"
	"yapi.run/cli/internal/cli/commands"
)

func main() {
	outputDir := "./web/app/_docs"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	// Build command tree without handlers (for doc generation only)
	rootCmd := commands.BuildRoot(nil, nil)

	// Custom link handler to strip .md extension for web routes
	linkHandler := func(name string) string {
		return strings.TrimSuffix(name, ".md")
	}

	if err := doc.GenMarkdownTreeCustom(rootCmd, outputDir, func(string) string { return "" }, linkHandler); err != nil {
		log.Fatalf("failed to generate docs: %v", err)
	}

	log.Printf("Generated %d docs in %s", len(rootCmd.Commands())+1, outputDir)
}

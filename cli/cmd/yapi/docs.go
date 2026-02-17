package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/docs"
)

// TODO: Consolidate cobra-generated command docs (--help) with manual topic docs
// in internal/docs/topics/ so there's a single source of truth for all documentation.
func docsE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return printTopicIndex()
	}
	return printTopic(args[0])
}

func printTopicIndex() error {
	fmt.Println("Available documentation topics:")
	fmt.Println()
	for _, t := range docs.List() {
		fmt.Printf("  yapi docs %-15s %s\n", t.Name, t.Summary)
	}
	fmt.Println()
	fmt.Println("Run 'yapi docs <topic>' to read a topic.")
	return nil
}

func printTopic(name string) error {
	content, err := docs.Get(name)
	if err != nil {
		// Try fuzzy suggestion
		if suggestion := docs.Suggest(name); suggestion != "" {
			fmt.Fprintf(os.Stderr, "Unknown topic %q. Did you mean %q?\n\n", name, suggestion)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown topic %q.\n\n", name)
		}
		fmt.Fprintf(os.Stderr, "Available topics: %s\n", strings.Join(docs.TopicNames(), ", "))
		return fmt.Errorf("unknown topic %q", name)
	}
	rendered, err := docs.Render(content)
	if err != nil {
		return fmt.Errorf("rendering docs: %w", err)
	}
	fmt.Print(rendered)
	return nil
}

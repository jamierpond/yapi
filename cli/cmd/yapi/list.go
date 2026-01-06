package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/tui"
)

func listE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Determine search directory
	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	// If no directory specified, try git-based discovery
	// If directory specified, use file walk
	var yapiFiles []string
	var err error

	if len(args) == 0 {
		// Use tui.FindConfigFiles to get git-tracked yapi files
		yapiFiles, err = tui.FindConfigFiles()
		if err != nil {
			// Fall back to file walk if not in git repo
			yapiFiles, err = findAllYapiFiles(searchDir)
		}
	} else {
		// Directory specified - use file walk
		yapiFiles, err = findAllYapiFiles(searchDir)
	}

	if err != nil {
		return fmt.Errorf("failed to find yapi files: %w", err)
	}

	if len(yapiFiles) == 0 {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No yapi config files found"))
		}
		return nil
	}

	// Sort files alphabetically
	sort.Strings(yapiFiles)

	// Output as JSON or text
	if jsonOutput {
		type fileEntry struct {
			Path string `json:"path"`
		}
		entries := make([]fileEntry, len(yapiFiles))
		for i, file := range yapiFiles {
			entries[i].Path = file
		}
		output, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(output))
	} else {
		// Text output
		fmt.Fprintf(os.Stderr, "%s\n\n", color.Accent(fmt.Sprintf("Found %d yapi config file(s):", len(yapiFiles))))
		for _, file := range yapiFiles {
			fmt.Println(file)
		}
	}

	return nil
}

// findAllYapiFiles finds all *.yapi, *.yapi.yml, *.yapi.yaml files in the given directory (excluding yapi.config.yml/yaml)
func findAllYapiFiles(dir string) ([]string, error) {
	var yapiFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		ext := filepath.Ext(path)

		// Match *.yapi, *.yapi.yml or *.yapi.yaml (but not yapi.config.yml/yaml)
		if base != "yapi.config.yml" && base != "yapi.config.yaml" {
			if strings.HasSuffix(base, ".yapi.yml") || strings.HasSuffix(base, ".yapi.yaml") {
				yapiFiles = append(yapiFiles, path)
			} else if ext == ".yapi" {
				yapiFiles = append(yapiFiles, path)
			}
		}
		return nil
	})

	return yapiFiles, err
}

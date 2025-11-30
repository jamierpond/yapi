package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ktr0731/go-fuzzyfinder"
)

// FindConfigFile searches for .yapi.yml files and lets the user select one.
// It starts searching from the current directory and goes up to 5 parent directories.
func FindConfigFile() (string, error) {
	var configFiles []string
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Search for .yapi.yml files in current and parent directories
	dir := currentDir
	for i := 0; i < 5; i++ { // Limit search to current + 4 parent directories
		files, err := filepath.Glob(filepath.Join(dir, "*.yapi.yml"))
		if err != nil {
			return "", fmt.Errorf("failed to glob for config files in %s: %w", dir, err)
		}
		for _, file := range files {
			// Store relative path to CWD
			relPath, err := filepath.Rel(currentDir, file)
			if err != nil {
				relPath = file // Fallback to absolute path if relative fails
			}
			configFiles = append(configFiles, relPath)
		}

		// Also search for config files in the adjacent 'examples' directory
		if dir == currentDir { // Only search 'examples' relative to the initial CWD
			exampleFiles, err := filepath.Glob(filepath.Join(filepath.Dir(currentDir), "examples", "*.yapi.yml"))
			if err != nil {
				return "", fmt.Errorf("failed to glob for example config files: %w", err)
			}
			for _, file := range exampleFiles {
				relPath, err := filepath.Rel(currentDir, file)
				if err != nil {
					relPath = file
				}
				configFiles = append(configFiles, relPath)
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir { // Reached root or unable to go up
			break
		}
		dir = parentDir
	}

	if len(configFiles) == 0 {
		return "", fmt.Errorf("no .yapi.yml files found in current or parent directories")
	}

	sort.Strings(configFiles)

	idx, err := fuzzyfinder.Find(
		configFiles,
		func(i int) string { return configFiles[i] },
	)
	if err != nil {
		if err == fuzzyfinder.ErrAbort { // User cancelled
			return "", fmt.Errorf("configuration selection aborted by user")
		}
		return "", fmt.Errorf("failed to find config file interactively: %w", err)
	}

	return configFiles[idx], nil
}

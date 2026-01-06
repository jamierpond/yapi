package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
)

func (app *rootCommand) testE(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	envName, _ := cmd.Flags().GetString("env")
	all, _ := cmd.Flags().GetBool("all")
	parallel, _ := cmd.Flags().GetInt("parallel")

	if parallel < 1 {
		return fmt.Errorf("parallel must be at least 1")
	}

	// Determine search directory
	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	// Find all test files
	testFiles, err := findTestFiles(searchDir, all)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	if len(testFiles) == 0 {
		if all {
			fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.yapi, *.yapi.yml, or *.yapi.yaml files found"))
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.test.yapi, *.test.yapi.yml, or *.test.yapi.yaml files found"))
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s\n\n", color.Accent(fmt.Sprintf("Running %d test(s)...", len(testFiles))))

	// Run each test and collect results
	type testResult struct {
		file   string
		index  int
		passed bool
		err    error
	}

	// Create channels and wait group for parallel execution
	results := make(chan testResult, len(testFiles))
	semaphore := make(chan struct{}, parallel)
	var wg sync.WaitGroup

	// Launch all tests in parallel (controlled by semaphore)
	for i, testFile := range testFiles {
		wg.Add(1)
		go func(idx int, filePath string) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			relPath, _ := filepath.Rel(searchDir, filePath)
			if verbose {
				fmt.Fprintf(os.Stderr, "%s %s\n", color.Dim(fmt.Sprintf("[%d/%d]", idx+1, len(testFiles))), relPath)
			}

			// Run the test file
			err := app.executeRunE(runContext{path: filePath, strict: true, envName: envName})

			result := testResult{
				file:   relPath,
				index:  idx,
				passed: err == nil,
				err:    err,
			}
			results <- result

			if err == nil {
				if !verbose {
					fmt.Fprintf(os.Stderr, "%s ", color.Green("OK"))
				} else {
					fmt.Fprintf(os.Stderr, "  %s\n\n", color.Green("PASS"))
				}
			} else {
				if !verbose {
					fmt.Fprintf(os.Stderr, "%s ", color.Red("X"))
				} else {
					fmt.Fprintf(os.Stderr, "  %s %s\n\n", color.Red("FAIL"), color.Dim(err.Error()))
				}
			}
		}(i, testFile)
	}

	// Wait for all tests to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allResults []testResult
	passCount := 0
	for result := range results {
		allResults = append(allResults, result)
		if result.passed {
			passCount++
		}
	}

	// Sort results by original index to maintain order in summary
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].index < allResults[j].index
	})

	if !verbose {
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\n")
	if passCount == len(allResults) {
		fmt.Fprintf(os.Stderr, "%s\n", color.Green(fmt.Sprintf("All %d test(s) passed", passCount)))
		return nil
	}

	failCount := len(allResults) - passCount
	fmt.Fprintf(os.Stderr, "%s\n", color.Red(fmt.Sprintf("%d of %d test(s) failed", failCount, len(allResults))))

	// List failed tests
	fmt.Fprintf(os.Stderr, "\n%s\n", color.Red("Failed tests:"))
	for _, r := range allResults {
		if !r.passed {
			fmt.Fprintf(os.Stderr, "  %s %s\n", color.Red("X"), r.file)
			if r.err != nil && verbose {
				fmt.Fprintf(os.Stderr, "    %s\n", color.Dim(r.err.Error()))
			}
		}
	}

	return fmt.Errorf("%d test(s) failed", failCount)
}

// findTestFiles recursively finds test files in the given directory.
// If all is true, finds all *.yapi, *.yapi.yml, *.yapi.yaml files.
// If all is false, finds only *.test.yapi, *.test.yapi.yml, *.test.yapi.yaml files.
func findTestFiles(dir string, all bool) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		ext := filepath.Ext(path)

		if all {
			// Match *.yapi, *.yapi.yml or *.yapi.yaml (but not yapi.config.yml/yaml)
			if base != "yapi.config.yml" && base != "yapi.config.yaml" {
				if strings.HasSuffix(base, ".yapi.yml") || strings.HasSuffix(base, ".yapi.yaml") {
					testFiles = append(testFiles, path)
				} else if ext == ".yapi" {
					testFiles = append(testFiles, path)
				}
			}
		} else {
			// Match *.test.yapi, *.test.yapi.yml or *.test.yapi.yaml
			if strings.HasSuffix(base, ".test.yapi.yml") || strings.HasSuffix(base, ".test.yapi.yaml") {
				testFiles = append(testFiles, path)
			} else if strings.HasSuffix(base, ".test.yapi") && ext == ".yapi" {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})

	return testFiles, err
}

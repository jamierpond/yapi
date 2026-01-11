package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/healthcheck"
	"yapi.run/cli/internal/process"
)

// testOptions holds parsed test command options.
type testOptions struct {
	verbose        bool
	envName        string
	all            bool
	parallel       int
	noStart        bool
	startOverride  string
	waitOnOverride []string
	waitTimeout    time.Duration
	searchDir      string
}

func (app *rootCommand) testE(cmd *cobra.Command, args []string) error {
	opts := parseTestFlags(cmd, args)

	if opts.parallel < 1 {
		return fmt.Errorf("parallel must be at least 1")
	}

	// Load and apply project config
	testConfig := loadTestConfig(opts.searchDir)
	applyTestConfigDefaults(&opts, testConfig, args)

	// Start server if configured
	proc, err := maybeStartServer(opts, testConfig)
	if err != nil {
		return err
	}
	if proc != nil {
		// Set up signal handler to clean up server on Ctrl+C
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			_ = proc.Stop()
			os.Exit(130) // 128 + SIGINT(2)
		}()
		defer func() {
			signal.Stop(sigChan)
			_ = proc.Stop()
		}()
	}

	// Find and run tests
	return app.runTestFiles(opts)
}

func parseTestFlags(cmd *cobra.Command, args []string) testOptions {
	verbose, _ := cmd.Flags().GetBool("verbose")
	envName, _ := cmd.Flags().GetString("env")
	all, _ := cmd.Flags().GetBool("all")
	parallel, _ := cmd.Flags().GetInt("parallel")
	noStart, _ := cmd.Flags().GetBool("no-start")
	startOverride, _ := cmd.Flags().GetString("start")
	waitOnOverride, _ := cmd.Flags().GetStringSlice("wait-on")
	waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")

	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	return testOptions{
		verbose:        verbose,
		envName:        envName,
		all:            all,
		parallel:       parallel,
		noStart:        noStart,
		startOverride:  startOverride,
		waitOnOverride: waitOnOverride,
		waitTimeout:    waitTimeout,
		searchDir:      searchDir,
	}
}

func loadTestConfig(searchDir string) *config.TestConfig {
	projectRoot, err := config.FindProjectRoot(searchDir)
	if err != nil {
		return nil
	}
	project, err := config.LoadProject(projectRoot)
	if err != nil {
		return nil
	}
	return project.Test
}

func applyTestConfigDefaults(opts *testOptions, testConfig *config.TestConfig, args []string) {
	if testConfig == nil {
		return
	}
	if opts.parallel == 1 && testConfig.Parallel > 0 {
		opts.parallel = testConfig.Parallel
	}
	if !opts.verbose && testConfig.Verbose {
		opts.verbose = testConfig.Verbose
	}
	if !opts.all && testConfig.All {
		opts.all = testConfig.All
	}
	if testConfig.Directory != "" && len(args) == 0 {
		opts.searchDir = testConfig.Directory
	}
}

func maybeStartServer(opts testOptions, testConfig *config.TestConfig) (*process.ManagedProcess, error) {
	if opts.noStart {
		return nil, nil
	}

	startCmd, waitOn, timeout := resolveServerConfig(opts, testConfig)
	if startCmd == "" {
		return nil, nil
	}

	fmt.Fprintf(os.Stderr, "%s %s\n", color.Accent("Starting server:"), startCmd)

	proc, err := process.Start(context.Background(), startCmd, opts.verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	if len(waitOn) > 0 {
		fmt.Fprintf(os.Stderr, "%s %s (timeout: %s)\n", color.Accent("Waiting for health:"), strings.Join(waitOn, ", "), timeout)

		if err := healthcheck.WaitForHealth(context.Background(), waitOn, timeout); err != nil {
			_ = proc.Stop()
			return nil, fmt.Errorf("health check failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "%s\n\n", color.Green("Server healthy!"))
	}

	return proc, nil
}

func resolveServerConfig(opts testOptions, testConfig *config.TestConfig) (string, []string, time.Duration) {
	startCmd := opts.startOverride
	waitOn := opts.waitOnOverride
	timeout := opts.waitTimeout

	if startCmd == "" && testConfig != nil {
		startCmd = testConfig.Start
	}
	if len(waitOn) == 0 && testConfig != nil {
		waitOn = testConfig.WaitOn
	}
	if timeout == 0 && testConfig != nil && testConfig.Timeout != "" {
		if parsed, err := time.ParseDuration(testConfig.Timeout); err == nil {
			timeout = parsed
		}
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return startCmd, waitOn, timeout
}

func (app *rootCommand) runTestFiles(opts testOptions) error {
	testFiles, err := findTestFiles(opts.searchDir, opts.all)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	if len(testFiles) == 0 {
		printNoTestsFound(opts.all)
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s\n\n", color.Accent(fmt.Sprintf("Running %d test(s)...", len(testFiles))))

	results := app.executeTests(testFiles, opts)
	return printTestResults(results, opts.verbose)
}

func printNoTestsFound(all bool) {
	if all {
		fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.yapi, *.yapi.yml, or *.yapi.yaml files found"))
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No *.test.yapi, *.test.yapi.yml, or *.test.yapi.yaml files found"))
	}
}

type testResult struct {
	file   string
	index  int
	passed bool
	err    error
}

func (app *rootCommand) executeTests(testFiles []string, opts testOptions) []testResult {
	results := make(chan testResult, len(testFiles))
	semaphore := make(chan struct{}, opts.parallel)
	var wg sync.WaitGroup

	for i, testFile := range testFiles {
		wg.Add(1)
		go func(idx int, filePath string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			relPath, _ := filepath.Rel(opts.searchDir, filePath)
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "%s %s\n", color.Dim(fmt.Sprintf("[%d/%d]", idx+1, len(testFiles))), relPath)
			}

			err := app.executeRunE(runContext{path: filePath, strict: true, envName: opts.envName})

			results <- testResult{file: relPath, index: idx, passed: err == nil, err: err}

			printTestProgress(err, opts.verbose)
		}(i, testFile)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []testResult
	for result := range results {
		allResults = append(allResults, result)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].index < allResults[j].index
	})

	return allResults
}

func printTestProgress(err error, verbose bool) {
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
}

func printTestResults(results []testResult, verbose bool) error {
	passCount := 0
	for _, r := range results {
		if r.passed {
			passCount++
		}
	}

	if !verbose {
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "\n")
	if passCount == len(results) {
		fmt.Fprintf(os.Stderr, "%s\n", color.Green(fmt.Sprintf("All %d test(s) passed", passCount)))
		return nil
	}

	failCount := len(results) - passCount
	fmt.Fprintf(os.Stderr, "%s\n", color.Red(fmt.Sprintf("%d of %d test(s) failed", failCount, len(results))))

	fmt.Fprintf(os.Stderr, "\n%s\n", color.Red("Failed tests:"))
	for _, r := range results {
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
			if base != "yapi.config.yml" && base != "yapi.config.yaml" {
				if strings.HasSuffix(base, ".yapi.yml") || strings.HasSuffix(base, ".yapi.yaml") || ext == ".yapi" {
					testFiles = append(testFiles, path)
				}
			}
		} else {
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

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/validation"
)

// stressTestResult represents the result of a single stress test request
type stressTestResult struct {
	duration time.Duration
	err      error
}

// promptStressTestConfirmation shows a confirmation prompt for stress testing
func (app *rootCommand) promptStressTestConfirmation(filePath, envName string, parallel, numRequests int, duration time.Duration, useDuration bool) error {
	// Load project and environment
	projEnv, err := loadProjectAndEnv(filePath, envName, true)
	if err != nil {
		return err
	}

	// Read config file
	configData, err := os.ReadFile(filePath) // #nosec G304 -- filePath is validated user-provided config file path
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Build resolver with environment variables
	var analysis *validation.Analysis
	if projEnv != nil && projEnv.project != nil && projEnv.envVars != nil {
		// Temporarily override the default environment
		originalDefault := projEnv.project.DefaultEnvironment
		projEnv.project.DefaultEnvironment = projEnv.envName
		analysis, err = validation.Analyze(string(configData), validation.AnalyzeOptions{
			FilePath:    filePath,
			Project:     projEnv.project,
			ProjectRoot: projEnv.projectRoot,
		})
		// Restore original default
		projEnv.project.DefaultEnvironment = originalDefault
	} else {
		analysis, err = validation.Analyze(string(configData), validation.AnalyzeOptions{FilePath: filePath})
	}
	if err != nil {
		return fmt.Errorf("failed to analyze config: %w", err)
	}

	// Get the resolved URL
	var targetURL string
	if analysis.Request != nil && analysis.Request.URL != "" {
		targetURL = analysis.Request.URL
	} else {
		targetURL = "<unknown>"
	}

	// Show confirmation prompt
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("Warning: Stress Test Confirmation"))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s %s\n", color.Dim("Target:"), color.Cyan(targetURL))
	fmt.Fprintf(os.Stderr, "  %s %d threads\n", color.Dim("Threads:"), parallel)
	if useDuration {
		fmt.Fprintf(os.Stderr, "  %s %v\n", color.Dim("Duration:"), duration)
		fmt.Fprintf(os.Stderr, "  %s unlimited (duration-based)\n", color.Dim("Requests:"))
	} else {
		fmt.Fprintf(os.Stderr, "  %s %d total (%d per thread)\n", color.Dim("Requests:"), numRequests, numRequests/parallel)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Are you sure you want to continue? [y/N]: ")

	var response string
	_, err = fmt.Scanln(&response)
	if err != nil || (response != "y" && response != "Y" && response != "yes" && response != "YES") {
		fmt.Fprintf(os.Stderr, "\n%s\n", color.Yellow("Stress test cancelled"))
		return fmt.Errorf("cancelled")
	}
	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

// printStressTestResults calculates and prints stress test statistics
func printStressTestResults(allResults []stressTestResult, startTime, stopTime time.Time, parallel, numRequests int, useDuration bool, duration time.Duration) error {
	if len(allResults) == 0 {
		return fmt.Errorf("no requests completed")
	}

	var durations []time.Duration
	successCount := 0
	failCount := 0
	var totalDuration time.Duration

	for _, r := range allResults {
		durations = append(durations, r.duration)
		totalDuration += r.duration
		if r.err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	// Sort durations for percentile calculations
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	totalTime := stopTime.Sub(startTime)
	reqsPerSec := float64(len(allResults)) / totalTime.Seconds()

	// Calculate percentiles
	p50 := durations[len(durations)*50/100]
	p90 := durations[len(durations)*90/100]
	p95 := durations[len(durations)*95/100]
	p99 := durations[len(durations)*99/100]
	minDuration := durations[0]
	maxDuration := durations[len(durations)-1]
	avgDuration := totalDuration / time.Duration(len(durations))

	// Print results
	fmt.Fprintf(os.Stderr, "%s\n", color.Accent("Results:"))
	fmt.Fprintf(os.Stderr, "\n")

	// Configuration summary
	if useDuration {
		fmt.Fprintf(os.Stderr, "  %s\n", color.Dim(fmt.Sprintf("%d threads, %v duration, %d total requests", parallel, duration, len(allResults))))
	} else {
		requestsPerThread := numRequests / parallel
		remainder := numRequests % parallel
		if remainder > 0 {
			fmt.Fprintf(os.Stderr, "  %s\n", color.Dim(fmt.Sprintf("%d threads, ~%d requests/thread, %d total requests", parallel, requestsPerThread, numRequests)))
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n", color.Dim(fmt.Sprintf("%d threads, %d requests/thread, %d total requests", parallel, requestsPerThread, numRequests)))
		}
	}
	fmt.Fprintf(os.Stderr, "  %s\n", color.Dim(fmt.Sprintf("Completed in %.2f seconds", totalTime.Seconds())))
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "  %s\n", color.Green(fmt.Sprintf("Success: %d (%.1f%%)", successCount, float64(successCount)*100/float64(len(allResults)))))
	if failCount > 0 {
		fmt.Fprintf(os.Stderr, "  %s\n", color.Red(fmt.Sprintf("Failed: %d (%.1f%%)", failCount, float64(failCount)*100/float64(len(allResults)))))
	} else {
		fmt.Fprintf(os.Stderr, "  %s\n", color.Dim(fmt.Sprintf("Failed: 0 (0.0%%)")))
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s\n", color.Accent("Throughput:"))
	fmt.Fprintf(os.Stderr, "    %.2f requests/second\n", reqsPerSec)
	fmt.Fprintf(os.Stderr, "    %.2f ms/request (avg)\n", float64(avgDuration.Microseconds())/1000.0)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s\n", color.Accent("Latency:"))
	fmt.Fprintf(os.Stderr, "    Min:  %v\n", minDuration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "    Avg:  %v\n", avgDuration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "    Max:  %v\n", maxDuration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s\n", color.Accent("Percentiles:"))
	fmt.Fprintf(os.Stderr, "    50%%:  %v\n", p50.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "    90%%:  %v\n", p90.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "    95%%:  %v\n", p95.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "    99%%:  %v\n", p99.Round(time.Millisecond))

	if failCount > 0 {
		return fmt.Errorf("%d requests failed", failCount)
	}

	return nil
}

func (app *rootCommand) stressE(cmd *cobra.Command, args []string) error {
	parallel, _ := cmd.Flags().GetInt("parallel")
	numRequests, _ := cmd.Flags().GetInt("num-requests")
	durationStr, _ := cmd.Flags().GetString("duration")
	envName, _ := cmd.Flags().GetString("env")
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	if parallel < 1 {
		return fmt.Errorf("parallel must be at least 1")
	}

	// Handle TUI file selection when no args provided
	filePath, _, err := selectConfigFile(args, "stress")
	if err != nil {
		return err
	}

	// Parse duration if provided
	var duration time.Duration
	var useDuration bool
	if durationStr != "" {
		var err error
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		useDuration = true
	} else if numRequests < 1 {
		return fmt.Errorf("num-requests must be at least 1")
	}

	// Show confirmation prompt
	if !skipConfirm {
		if err := app.promptStressTestConfirmation(filePath, envName, parallel, numRequests, duration, useDuration); err != nil {
			return nil //nolint:nilerr // User cancelled, exit without error
		}
	}

	// Print header
	fmt.Fprintf(os.Stderr, "%s\n", color.Accent("yapi stress test"))
	fmt.Fprintf(os.Stderr, "%s\n", color.Dim("File: "+filePath))
	if useDuration {
		fmt.Fprintf(os.Stderr, "%s\n", color.Dim(fmt.Sprintf("Duration: %v, Concurrency: %d", duration, parallel)))
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", color.Dim(fmt.Sprintf("Requests: %d, Concurrency: %d", numRequests, parallel)))
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Statistics tracking
	results := make(chan stressTestResult, parallel)
	var wg sync.WaitGroup

	startTime := time.Now()
	var stopTime time.Time

	// Worker function
	worker := func(requestCount *int64) {
		defer wg.Done()
		for {
			// Check if we should stop
			if useDuration {
				if time.Since(startTime) >= duration {
					return
				}
			} else {
				if atomic.AddInt64(requestCount, 1) > int64(numRequests) {
					return
				}
			}

			// Execute request (returnErrors: true ensures errors are captured for counting)
			reqStart := time.Now()
			err := app.executeRunE(runContext{path: filePath, strict: false, returnErrors: true, envName: envName})
			reqDuration := time.Since(reqStart)

			results <- stressTestResult{duration: reqDuration, err: err}
		}
	}

	// Start workers
	var requestCount int64
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go worker(&requestCount)
	}

	// Collect results in a separate goroutine
	var allResults []stressTestResult
	done := make(chan bool)
	go func() {
		for result := range results {
			allResults = append(allResults, result)
			// Print progress
			if len(allResults)%10 == 0 || (!useDuration && len(allResults) == numRequests) {
				elapsed := time.Since(startTime)
				rps := float64(len(allResults)) / elapsed.Seconds()
				fmt.Fprintf(os.Stderr, "\r%s %d requests in %v (%.2f req/s)",
					color.Dim("Progress:"), len(allResults), elapsed.Round(time.Millisecond), rps)
			}
		}
		done <- true
	}()

	// Wait for all workers to finish
	wg.Wait()
	stopTime = time.Now()
	close(results)
	<-done

	fmt.Fprintf(os.Stderr, "\r%s\n\n", strings.Repeat(" ", 80)) // Clear progress line

	// Print results and statistics
	return printStressTestResults(allResults, startTime, stopTime, parallel, numRequests, useDuration, duration)
}

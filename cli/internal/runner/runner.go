// Package runner executes API requests and chains.
package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/filter"
	"yapi.run/cli/internal/vars"
)

const (
	defaultPollPeriod  = 1 * time.Second
	defaultPollTimeout = 30 * time.Second
)

// Result holds the output of a yapi execution
type Result struct {
	Body        string
	ContentType string
	StatusCode  int
	Warnings    []string
	RequestURL  string        // The full constructed URL (HTTP/GraphQL only)
	Duration    time.Duration // Time taken for the request
	BodyLines   int
	BodyChars   int
	BodyBytes   int
	Headers     map[string]string // Response headers
	OutputFile  string            // Path where output was saved (if output_file was specified)
}

// Options for execution
type Options struct {
	URLOverride    string
	NoColor        bool
	BinaryOutput   bool
	Insecure       bool
	Verbose        bool              // Show resolved request/response details during execution
	EnvOverrides   map[string]string // Environment variables from project config
	ProjectRoot    string            // Path to project root (for validation)
	ProjectEnv     string            // Selected environment name (for validation)
	ConfigFilePath string            // Path to the yapi config file (for relative output_file resolution)
	StrictEnv      bool              // Strict env mode: error on missing env files, no OS env fallback
}

// Run executes a yapi request and returns the result.
func Run(ctx context.Context, exec executor.TransportFunc, req *domain.Request, warnings []string, opts Options) (*Result, error) {
	if opts.Insecure {
		if req.Metadata == nil {
			req.Metadata = make(map[string]string)
		}
		req.Metadata["insecure"] = "true"
	}

	// Apply URL override
	if opts.URLOverride != "" {
		req.URL = opts.URLOverride
	}

	// Execute the request
	resp, err := exec(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	body := string(bodyBytes)

	// Apply JQ filter if specified
	if jqFilter, ok := req.Metadata["jq_filter"]; ok && jqFilter != "" {
		body, err = filter.ApplyJQ(body, jqFilter)
		if err != nil {
			return nil, fmt.Errorf("jq filter failed: %w", err)
		}
		resp.Headers["Content-Type"] = "application/json"
	}

	// Write to output file if specified
	var savedOutputFile string
	if outputFile, ok := req.Metadata["output_file"]; ok && outputFile != "" {
		// Resolve relative paths against the config file directory
		if !filepath.IsAbs(outputFile) && opts.ConfigFilePath != "" {
			outputFile = filepath.Join(filepath.Dir(opts.ConfigFilePath), outputFile)
		}
		// Create parent directories if they don't exist
		if dir := filepath.Dir(outputFile); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0750); err != nil {
				return nil, fmt.Errorf("failed to create output directory '%s': %w", dir, err)
			}
		}
		if err := os.WriteFile(outputFile, []byte(body), 0600); err != nil {
			return nil, fmt.Errorf("failed to write output file '%s': %w", outputFile, err)
		}
		savedOutputFile = outputFile
	}

	bodyLines := strings.Count(body, "\n") + 1
	bodyChars := len(body)
	bodyBytesLen := len(bodyBytes)

	return &Result{
		Body:        body,
		ContentType: resp.Headers["Content-Type"],
		StatusCode:  resp.StatusCode,
		Warnings:    warnings,
		RequestURL:  req.URL,
		Duration:    resp.Duration,
		BodyLines:   bodyLines,
		BodyChars:   bodyChars,
		BodyBytes:   bodyBytesLen,
		Headers:     resp.Headers,
		OutputFile:  savedOutputFile,
	}, nil
}

// PollResult extends Result with polling-specific information
type PollResult struct {
	*Result
	Attempts int           // Number of attempts made
	Elapsed  time.Duration // Total time spent polling
}

// RunWithPolling executes a request repeatedly until conditions are met or timeout expires.
func RunWithPolling(ctx context.Context, exec executor.TransportFunc, req *domain.Request, waitFor *config.WaitFor, warnings []string, opts Options, envVars map[string]string) (*PollResult, error) {
	pollCfg, err := parseWaitForConfig(waitFor)
	if err != nil {
		return nil, err
	}

	// Read body bytes once so we can recreate the reader for each attempt.
	// io.Reader is consumed on first read, so without this, subsequent
	// polling attempts would send empty bodies.
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	startTime := time.Now()
	deadline := startTime.Add(pollCfg.timeout)
	attempt := 0
	jqVars := prepareJQVars(envVars)

	for {
		attempt++

		// Check if we've exceeded the timeout before making another attempt
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("wait_for timeout after %v (%d attempts made)", time.Since(startTime).Round(time.Millisecond), attempt-1)
		}

		// Reset body reader for this attempt
		if bodyBytes != nil {
			req.Body = bytes.NewReader(bodyBytes)
		}

		// Execute the request
		result, err := Run(ctx, exec, req, warnings, opts)
		if err != nil {
			// Request failed - check if we should retry or fail
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("wait_for timeout after %v (%d attempts made): last error: %w", time.Since(startTime).Round(time.Millisecond), attempt, err)
			}
			waitDuration := pollCfg.getWaitDuration(attempt)
			fmt.Fprintf(os.Stderr, "[wait_for] Attempt %d failed: %v, retrying in %v...\n", attempt, err, waitDuration)
			if !waitForDuration(ctx, waitDuration, deadline) {
				return nil, fmt.Errorf("wait_for timeout after %v (%d attempts made)", time.Since(startTime).Round(time.Millisecond), attempt)
			}
			continue
		}

		// Check if all "until" assertions pass
		allPassed := true
		for _, assertion := range waitFor.Until {
			outcome := evalAssertion(result.Body, assertion, jqVars)
			if !outcome.passed {
				allPassed = false
				break
			}
		}

		if allPassed {
			return &PollResult{
				Result:   result,
				Attempts: attempt,
				Elapsed:  time.Since(startTime),
			}, nil
		}

		// Assertions didn't pass - wait and retry
		waitDuration := pollCfg.getWaitDuration(attempt)
		fmt.Fprintf(os.Stderr, "[wait_for] Attempt %d: conditions not met, retrying in %v...\n", attempt, waitDuration)
		if !waitForDuration(ctx, waitDuration, deadline) {
			return nil, fmt.Errorf("wait_for timeout after %v (%d attempts made): conditions never met", time.Since(startTime).Round(time.Millisecond), attempt)
		}
	}
}

// pollConfig holds parsed polling configuration
type pollConfig struct {
	timeout    time.Duration
	period     time.Duration // Fixed period (if set)
	backoff    *backoffConfig
	useBackoff bool
}

// backoffConfig holds parsed backoff configuration
type backoffConfig struct {
	seed       time.Duration
	multiplier float64
}

// getWaitDuration returns the wait duration for the given attempt number
func (p *pollConfig) getWaitDuration(attempt int) time.Duration {
	if !p.useBackoff {
		return p.period
	}
	// Exponential backoff: seed * multiplier^(attempt-1), capped by the overall timeout.
	// attempt 1 = seed, attempt 2 = seed*multiplier, attempt 3 = seed*multiplier^2, etc.
	multiplied := float64(p.backoff.seed)
	maxWait := float64(p.timeout)

	for i := 1; i < attempt; i++ {
		multiplied *= p.backoff.multiplier
		if multiplied >= maxWait {
			return p.timeout
		}
	}
	return time.Duration(multiplied)
}

// parseWaitForConfig extracts polling configuration from WaitFor config
func parseWaitForConfig(waitFor *config.WaitFor) (*pollConfig, error) {
	// Validate until is non-empty
	if len(waitFor.Until) == 0 {
		return nil, fmt.Errorf("wait_for.until is required and must have at least one assertion")
	}

	// Validate period and backoff are mutually exclusive
	hasPeriod := waitFor.Period != ""
	hasBackoff := waitFor.Backoff != nil
	if hasPeriod && hasBackoff {
		return nil, fmt.Errorf("wait_for.period and wait_for.backoff are mutually exclusive")
	}
	if !hasPeriod && !hasBackoff {
		return nil, fmt.Errorf("wait_for requires either period or backoff to be specified")
	}

	cfg := &pollConfig{
		timeout: defaultPollTimeout,
	}

	// Parse timeout
	if waitFor.Timeout != "" {
		timeout, err := time.ParseDuration(waitFor.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid wait_for.timeout '%s': %w", waitFor.Timeout, err)
		}
		cfg.timeout = timeout
	}

	// Parse period or backoff
	if hasBackoff {
		cfg.useBackoff = true
		seed, err := time.ParseDuration(waitFor.Backoff.Seed)
		if err != nil {
			return nil, fmt.Errorf("invalid wait_for.backoff.seed '%s': %w", waitFor.Backoff.Seed, err)
		}
		multiplier := waitFor.Backoff.Multiplier
		if multiplier <= 1 {
			return nil, fmt.Errorf("wait_for.backoff.multiplier must be > 1, got %v", multiplier)
		}
		cfg.backoff = &backoffConfig{
			seed:       seed,
			multiplier: multiplier,
		}
	} else {
		period, err := time.ParseDuration(waitFor.Period)
		if err != nil {
			return nil, fmt.Errorf("invalid wait_for.period '%s': %w", waitFor.Period, err)
		}
		cfg.period = period
	}

	return cfg, nil
}

// waitForDuration waits for the duration or until deadline/context cancellation
// Returns false if the deadline would be exceeded
func waitForDuration(ctx context.Context, duration time.Duration, deadline time.Time) bool {
	// Don't wait if we'd exceed the deadline
	if time.Now().Add(duration).After(deadline) {
		return false
	}

	select {
	case <-time.After(duration):
		return true
	case <-ctx.Done():
		return false
	}
}

// ChainResult holds the output of a chain execution
type ChainResult struct {
	Results            []*Result            // Results from each step
	StepNames          []string             // Names of each step
	ExpectationResults []*ExpectationResult // Expectation results from each step
}

// ExecutorFactory is an interface for creating transport functions
type ExecutorFactory interface {
	Create(transport string) (executor.TransportFunc, error)
}

// RunChain executes a sequence of steps, merging each step with the base config
func RunChain(ctx context.Context, factory ExecutorFactory, base *config.ConfigV1, steps []config.ChainStep, opts Options) (*ChainResult, error) {
	chainCtx := NewChainContext(opts.EnvOverrides)
	chainResult := &ChainResult{
		Results:            make([]*Result, 0, len(steps)),
		StepNames:          make([]string, 0, len(steps)),
		ExpectationResults: make([]*ExpectationResult, 0, len(steps)),
	}

	for i, step := range steps {
		fmt.Fprintf(os.Stderr, "Running step %d: %s...\n", i+1, step.Name)

		// 1. Merge step with base config to get full config
		merged := base.Merge(step)

		// 1b. Warn about bare $word.word patterns that won't be substituted
		warnBareChainRefsInConfig(step.Name, &merged)

		// 2. Interpolate variables in the merged config
		interpolatedConfig, err := interpolateConfig(chainCtx, &merged)
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 2b. Verbose: show resolved request details
		if opts.Verbose {
			logResolvedConfig(i+1, step.Name, interpolatedConfig)
		}

		// 3. Handle Delay (wait before executing step)
		if interpolatedConfig.Delay != "" {
			d, err := time.ParseDuration(interpolatedConfig.Delay)
			if err != nil {
				return nil, fmt.Errorf("step '%s' invalid delay '%s': %w", step.Name, interpolatedConfig.Delay, err)
			}
			if d > 0 {
				fmt.Fprintf(os.Stderr, "[INFO] Delaying for %s...\n", d)
				select {
				case <-time.After(d):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}

		// 4. Convert to domain request (handles ALL transports: HTTP, TCP, gRPC, GraphQL)
		req, err := interpolatedConfig.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 5. Create executor for this step's transport
		exec, err := factory.Create(req.Metadata["transport"])
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 6. Execute (with polling if wait_for is configured)
		var result *Result
		if interpolatedConfig.WaitFor != nil {
			pollResult, err := RunWithPolling(ctx, exec, req, interpolatedConfig.WaitFor, []string{}, opts, opts.EnvOverrides)
			if err != nil {
				return nil, fmt.Errorf("step '%s' failed: %w", step.Name, err)
			}
			result = pollResult.Result
		} else {
			var err error
			result, err = Run(ctx, exec, req, []string{}, opts)
			if err != nil {
				return nil, fmt.Errorf("step '%s' failed: %w", step.Name, err)
			}
		}

		// 6b. Verbose: show response details
		if opts.Verbose {
			logStepResponse(i+1, step.Name, result)
		}

		// 7. Assert Expectations
		expectRes := CheckExpectationsWithEnv(step.Expect, result, opts.EnvOverrides)

		// 8. Store Result (including expectation result even if failed)
		chainCtx.AddResult(step.Name, result)
		chainResult.Results = append(chainResult.Results, result)
		chainResult.StepNames = append(chainResult.StepNames, step.Name)
		chainResult.ExpectationResults = append(chainResult.ExpectationResults, expectRes)

		if expectRes.Error != nil {
			return chainResult, fmt.Errorf("step '%s' assertion failed: %w", step.Name, expectRes.Error)
		}
	}

	return chainResult, nil
}

const verboseMaxLen = 1000

// truncateStr truncates s for verbose output at a valid UTF-8 boundary.
func truncateStr(s string) string {
	if len(s) <= verboseMaxLen {
		return s
	}
	// Walk back to avoid splitting a multi-byte UTF-8 character
	cut := verboseMaxLen
	for cut > 0 && s[cut-1]&0xC0 == 0x80 {
		cut--
	}
	if cut > 0 && s[cut-1]&0x80 != 0 {
		cut-- // skip the leading byte of the split character
	}
	return s[:cut] + fmt.Sprintf("... (%d bytes total)", len(s))
}

// logResolvedConfig prints the resolved request config to stderr for debugging.
func logResolvedConfig(stepNum int, stepName string, cfg *config.ConfigV1) {
	fmt.Fprintf(os.Stderr, "[VERBOSE] Step %d (%s) resolved request:\n", stepNum, stepName)
	if cfg.Method != "" || cfg.URL != "" {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   %s %s\n", cfg.Method, cfg.URL)
	}
	if cfg.Path != "" {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   Path: %s\n", cfg.Path)
	}
	for k, v := range cfg.Headers {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   Header: %s: %s\n", k, v)
	}
	if cfg.Body != nil {
		bodyJSON, err := json.Marshal(cfg.Body)
		if err == nil {
			fmt.Fprintf(os.Stderr, "[VERBOSE]   Body: %s\n", truncateStr(string(bodyJSON)))
		}
	}
	if cfg.JSON != "" {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   JSON: %s\n", truncateStr(cfg.JSON))
	}
	if cfg.Data != "" {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   Data: %s\n", truncateStr(cfg.Data))
	}
}

// logStepResponse prints response details to stderr for debugging.
func logStepResponse(stepNum int, stepName string, result *Result) {
	fmt.Fprintf(os.Stderr, "[VERBOSE] Step %d (%s) response:\n", stepNum, stepName)
	fmt.Fprintf(os.Stderr, "[VERBOSE]   Status: %d\n", result.StatusCode)
	fmt.Fprintf(os.Stderr, "[VERBOSE]   Duration: %s\n", result.Duration)
	if result.Body != "" {
		fmt.Fprintf(os.Stderr, "[VERBOSE]   Body: %s\n", truncateStr(result.Body))
	}
}

// warnBareChainRefsInConfig checks a config's string fields for bare $word.word patterns
// and prints warnings to stderr during chain execution.
func warnBareChainRefsInConfig(stepName string, cfg *config.ConfigV1) {
	check := func(field, value string) {
		refs := vars.FindBareRefs(value)
		for _, ref := range refs {
			fmt.Fprintf(os.Stderr, "[WARN] step '%s' %s: possible bare variable '%s' -- did you mean '${%s}'?\n",
				stepName, field, ref, ref[1:])
		}
	}

	check("url", cfg.URL)
	check("path", cfg.Path)
	check("json", cfg.JSON)
	check("data", cfg.Data)
	for k, v := range cfg.Headers {
		check(fmt.Sprintf("header '%s'", k), v)
	}
	for k, v := range cfg.Query {
		check(fmt.Sprintf("query '%s'", k), v)
	}
}

// interpolateConfig expands chain variables in a config
func interpolateConfig(chainCtx *ChainContext, cfg *config.ConfigV1) (*config.ConfigV1, error) {
	result := *cfg // Copy

	// Interpolate URL
	if result.URL != "" {
		expanded, err := chainCtx.ExpandVariables(result.URL)
		if err != nil {
			return nil, fmt.Errorf("url: %w", err)
		}
		result.URL = expanded
	}

	// Interpolate Path
	if result.Path != "" {
		expanded, err := chainCtx.ExpandVariables(result.Path)
		if err != nil {
			return nil, fmt.Errorf("path: %w", err)
		}
		result.Path = expanded
	}

	// Interpolate Headers
	if result.Headers != nil {
		newHeaders := make(map[string]string)
		for k, v := range result.Headers {
			expanded, err := chainCtx.ExpandVariables(v)
			if err != nil {
				return nil, fmt.Errorf("header '%s': %w", k, err)
			}
			newHeaders[k] = expanded
		}
		result.Headers = newHeaders
	}

	// Interpolate Query params
	if result.Query != nil {
		newQuery := make(map[string]string)
		for k, v := range result.Query {
			expanded, err := chainCtx.ExpandVariables(v)
			if err != nil {
				return nil, fmt.Errorf("query '%s': %w", k, err)
			}
			newQuery[k] = expanded
		}
		result.Query = newQuery
	}

	// Interpolate JSON
	if result.JSON != "" {
		expanded, err := chainCtx.ExpandVariables(result.JSON)
		if err != nil {
			return nil, fmt.Errorf("json: %w", err)
		}
		result.JSON = expanded
	}

	// Interpolate Data (TCP)
	if result.Data != "" {
		expanded, err := chainCtx.ExpandVariables(result.Data)
		if err != nil {
			return nil, fmt.Errorf("data: %w", err)
		}
		result.Data = expanded
	}

	// Interpolate Body
	if result.Body != nil {
		newBody, err := interpolateBody(chainCtx, result.Body)
		if err != nil {
			return nil, fmt.Errorf("body: %w", err)
		}
		result.Body = newBody
	}

	// Interpolate Variables (GraphQL)
	if result.Variables != nil {
		newVars, err := interpolateBody(chainCtx, result.Variables)
		if err != nil {
			return nil, fmt.Errorf("variables: %w", err)
		}
		result.Variables = newVars
	}

	// Interpolate Delay
	if result.Delay != "" {
		expanded, err := chainCtx.ExpandVariables(result.Delay)
		if err != nil {
			return nil, fmt.Errorf("delay: %w", err)
		}
		result.Delay = expanded
	}

	// Interpolate OutputFile
	if result.OutputFile != "" {
		expanded, err := chainCtx.ExpandVariables(result.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("output_file: %w", err)
		}
		result.OutputFile = expanded
	}

	return &result, nil
}

// interpolateBody recursively interpolates variables in body map
// It preserves types for pure variable references (e.g. $step.field returns int/bool, not string)
func interpolateBody(chainCtx *ChainContext, body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, nil
	}

	result := make(map[string]any)
	for k, v := range body {
		interpolated, err := interpolateValue(chainCtx, v)
		if err != nil {
			return nil, err
		}
		result[k] = interpolated
	}
	return result, nil
}

// interpolateValue recursively interpolates variables in any value
func interpolateValue(chainCtx *ChainContext, v any) (any, error) {
	switch val := v.(type) {
	case string:
		// First, try to resolve as a pure variable reference (preserves type)
		if rawVal, ok := chainCtx.ResolveVariableRaw(val); ok {
			return rawVal, nil
		}
		// Fall back to string interpolation
		return chainCtx.ExpandVariables(val)
	case map[string]any:
		return interpolateBody(chainCtx, val)
	case []any:
		result := make([]any, len(val))
		for i, elem := range val {
			interpolated, err := interpolateValue(chainCtx, elem)
			if err != nil {
				return nil, err
			}
			result[i] = interpolated
		}
		return result, nil
	default:
		return v, nil
	}
}

// AssertionResult holds the result of a single assertion
type AssertionResult struct {
	Expression    string
	Passed        bool
	Error         error
	ActualValue   string // The actual value from evaluation (for failed assertions)
	ExpectedValue string // The expected value (for failed assertions)
	LeftSide      string // The left side of comparison (e.g., ".id")
	Operator      string // The operator (e.g., "==", "!=", ">", etc.)
}

// formatAssertionError creates a detailed error message for a failed assertion
func formatAssertionError(detail *filter.AssertionDetail) string {
	if detail == nil {
		return "assertion failed"
	}

	// If we have detailed information about the comparison, use it
	if detail.LeftSide != "" && detail.Operator != "" {
		var operatorDesc string
		switch detail.Operator {
		case "==":
			operatorDesc = "to equal"
		case "!=":
			operatorDesc = "to not equal"
		case ">":
			operatorDesc = "to be greater than"
		case ">=":
			operatorDesc = "to be greater than or equal to"
		case "<":
			operatorDesc = "to be less than"
		case "<=":
			operatorDesc = "to be less than or equal to"
		default:
			operatorDesc = detail.Operator
		}

		// Build the error message with available information
		msg := fmt.Sprintf("assertion failed\n  Expected: %s %s %s",
			detail.LeftSide,
			operatorDesc,
			detail.ExpectedValue)

		if detail.ActualValue != "" {
			msg += fmt.Sprintf("\n  Actual:   %s = %s", detail.LeftSide, detail.ActualValue)
		}

		msg += fmt.Sprintf("\n  Expression: %s", detail.Expression)
		return msg
	}

	// Fallback to basic error message
	return fmt.Sprintf("assertion failed: %s", detail.Expression)
}

// ExpectationResult contains the results of running expectations
type ExpectationResult struct {
	StatusPassed     bool
	StatusChecked    bool
	AssertionsPassed int
	AssertionsTotal  int
	AssertionResults []AssertionResult
	Error            error
}

// AllPassed returns true if all expectations passed
func (e *ExpectationResult) AllPassed() bool {
	return e.Error == nil
}

// checkStatusMatch checks if the actual status code matches the expected status
func checkStatusMatch(expected any, actual int) bool {
	switch v := expected.(type) {
	case int:
		return actual == v
	case float64:
		return actual == int(v)
	case []any:
		for _, code := range v {
			switch c := code.(type) {
			case int:
				if c == actual {
					return true
				}
			case float64:
				if int(c) == actual {
					return true
				}
			}
		}
	}
	return false
}

// prepareJQVars creates the jq variables map from environment variables
func prepareJQVars(envVars map[string]string) map[string]any {
	if len(envVars) == 0 {
		return nil
	}
	envMap := make(map[string]any)
	for k, v := range envVars {
		envMap[k] = v
	}
	return map[string]any{"env": envMap}
}

// assertionOutcome holds the result of evaluating a single assertion
type assertionOutcome struct {
	result      AssertionResult
	passed      bool
	err         error
	errorDetail *filter.AssertionDetail
}

// evalAssertion evaluates a single assertion against JSON data
func evalAssertion(jsonData, assertion string, jqVars map[string]any) assertionOutcome {
	processedAssertion := strings.ReplaceAll(assertion, "env.", "$env.")

	var passed bool
	var detail *filter.AssertionDetail
	var err error

	if jqVars != nil {
		passed, detail, err = filter.EvalJQBoolWithDetailAndVars(jsonData, processedAssertion, jqVars)
	} else {
		passed, detail, err = filter.EvalJQBoolWithDetail(jsonData, processedAssertion)
	}

	ar := AssertionResult{
		Expression: assertion,
		Passed:     passed && err == nil,
		Error:      err,
	}
	if detail != nil {
		ar.ActualValue = detail.ActualValue
		ar.ExpectedValue = detail.ExpectedValue
		ar.LeftSide = detail.LeftSide
		ar.Operator = detail.Operator
	}

	return assertionOutcome{
		result:      ar,
		passed:      passed && err == nil,
		err:         err,
		errorDetail: detail,
	}
}

// CheckExpectationsWithEnv validates the response against expected values with environment variables
func CheckExpectationsWithEnv(expect config.Expectation, result *Result, envVars map[string]string) *ExpectationResult {
	totalAssertions := len(expect.Assert.Body) + len(expect.Assert.Headers)
	res := &ExpectationResult{
		AssertionsTotal:  totalAssertions,
		AssertionResults: make([]AssertionResult, 0, totalAssertions),
	}

	var firstError error
	var firstErrorDetail *filter.AssertionDetail

	// Status Check
	if expect.Status != nil {
		res.StatusChecked = true
		res.StatusPassed = checkStatusMatch(expect.Status, result.StatusCode)
		if !res.StatusPassed {
			firstError = fmt.Errorf("expected status %v, got %d", expect.Status, result.StatusCode)
		}
	}

	jqVars := prepareJQVars(envVars)

	// Body Assertions
	for _, assertion := range expect.Assert.Body {
		outcome := evalAssertion(result.Body, assertion, jqVars)
		res.AssertionResults = append(res.AssertionResults, outcome.result)

		//nolint:gocritic // ifElseChain: switch not suitable for boolean conditions
		if outcome.err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("assertion failed: %w", outcome.err)
			}
		} else if !outcome.passed {
			if firstError == nil {
				firstErrorDetail = outcome.errorDetail
			}
		} else {
			res.AssertionsPassed++
		}
	}

	// Header Assertions
	if len(expect.Assert.Headers) > 0 {
		headersJSON, err := json.Marshal(result.Headers)
		if err != nil {
			res.Error = fmt.Errorf("failed to marshal headers for assertions: %w", err)
			return res
		}

		for _, assertion := range expect.Assert.Headers {
			outcome := evalAssertion(string(headersJSON), assertion, jqVars)
			res.AssertionResults = append(res.AssertionResults, outcome.result)

			//nolint:gocritic // ifElseChain: switch not suitable for boolean conditions
			if outcome.err != nil {
				if firstError == nil {
					firstError = fmt.Errorf("header assertion failed: %w", outcome.err)
				}
			} else if !outcome.passed {
				if firstError == nil {
					firstErrorDetail = outcome.errorDetail
				}
			} else {
				res.AssertionsPassed++
			}
		}
	}

	// Set the first error encountered
	if firstError != nil {
		res.Error = firstError
	} else if firstErrorDetail != nil {
		res.Error = fmt.Errorf("%s", formatAssertionError(firstErrorDetail))
	}

	return res
}

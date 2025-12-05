package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/filter"
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
}

// Options for execution
type Options struct {
	URLOverride string
	NoColor     bool
}

// Run executes a yapi request and returns the result.
func Run(ctx context.Context, exec executor.TransportFunc, req *domain.Request, warnings []string, opts Options) (*Result, error) {
	// Apply URL override
	if opts.URLOverride != "" {
		req.URL = opts.URLOverride
	}

	// Execute the request
	resp, err := exec(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	}, nil
}

// ChainResult holds the output of a chain execution
type ChainResult struct {
	Results   []*Result // Results from each step
	StepNames []string  // Names of each step
}

// RunChain executes a sequence of steps, merging each step with the base config
func RunChain(ctx context.Context, factory *executor.Factory, base *config.ConfigV1, steps []config.ChainStep, opts Options) (*ChainResult, error) {
	chainCtx := NewChainContext()
	chainResult := &ChainResult{
		Results:   make([]*Result, 0, len(steps)),
		StepNames: make([]string, 0, len(steps)),
	}

	for i, step := range steps {
		fmt.Fprintf(os.Stderr, "Running step %d: %s...\n", i+1, step.Name)

		// 1. Merge step with base config to get full config
		merged := base.Merge(step)

		// 2. Interpolate variables in the merged config
		interpolatedConfig, err := interpolateConfig(chainCtx, &merged)
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 3. Convert to domain request (handles ALL transports: HTTP, TCP, gRPC, GraphQL)
		req, err := interpolatedConfig.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 4. Create executor for this step's transport
		exec, err := factory.Create(req.Metadata["transport"])
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 5. Execute
		result, err := Run(ctx, exec, req, []string{}, opts)
		if err != nil {
			return nil, fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}

		// 6. Assert Expectations
		expectRes := CheckExpectations(step.Expect, result)
		if expectRes.Error != nil {
			return nil, fmt.Errorf("step '%s' assertion failed: %w", step.Name, expectRes.Error)
		}

		// 7. Store Result
		chainCtx.AddResult(step.Name, result)
		chainResult.Results = append(chainResult.Results, result)
		chainResult.StepNames = append(chainResult.StepNames, step.Name)
	}

	return chainResult, nil
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

	return &result, nil
}

// interpolateBody recursively interpolates variables in body map
// It preserves types for pure variable references (e.g. $step.field returns int/bool, not string)
func interpolateBody(chainCtx *ChainContext, body map[string]interface{}) (map[string]interface{}, error) {
	if body == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for k, v := range body {
		switch val := v.(type) {
		case string:
			// First, try to resolve as a pure variable reference (preserves type)
			if rawVal, ok := chainCtx.ResolveVariableRaw(val); ok {
				result[k] = rawVal
			} else {
				// Fall back to string interpolation
				expanded, err := chainCtx.ExpandVariables(val)
				if err != nil {
					return nil, err
				}
				result[k] = expanded
			}
		case map[string]interface{}:
			nested, err := interpolateBody(chainCtx, val)
			if err != nil {
				return nil, err
			}
			result[k] = nested
		default:
			result[k] = v
		}
	}
	return result, nil
}

// ExpectationResult contains the results of running expectations
type ExpectationResult struct {
	StatusPassed     bool
	StatusChecked    bool
	AssertionsPassed int
	AssertionsTotal  int
	Error            error
}

// AllPassed returns true if all expectations passed
func (e *ExpectationResult) AllPassed() bool {
	return e.Error == nil
}

// CheckExpectations validates the response against expected values
func CheckExpectations(expect config.Expectation, result *Result) *ExpectationResult {
	res := &ExpectationResult{
		AssertionsTotal: len(expect.Assert),
	}

	// Status Check
	if expect.Status != nil {
		res.StatusChecked = true
		matched := false
		switch v := expect.Status.(type) {
		case int:
			if result.StatusCode == v {
				matched = true
			}
		case float64: // YAML often parses numbers as float64
			if result.StatusCode == int(v) {
				matched = true
			}
		case []interface{}: // YAML often parses arrays as []interface{}
			for _, code := range v {
				switch c := code.(type) {
				case int:
					if c == result.StatusCode {
						matched = true
					}
				case float64:
					if int(c) == result.StatusCode {
						matched = true
					}
				}
			}
		}
		res.StatusPassed = matched
		if !matched {
			res.Error = fmt.Errorf("expected status %v, got %d", expect.Status, result.StatusCode)
			return res
		}
	}

	// JQ Assertions
	for _, assertion := range expect.Assert {
		passed, err := filter.EvalJQBool(result.Body, assertion)
		if err != nil {
			res.Error = fmt.Errorf("assertion failed: %w", err)
			return res
		}
		if !passed {
			res.Error = fmt.Errorf("assertion failed: %s", assertion)
			return res
		}
		res.AssertionsPassed++
	}

	return res
}

package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/constants"
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
	RequestURL  string            // The full constructed URL (HTTP/GraphQL only)
	Duration    time.Duration     // Time taken for the request
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
func Run(ctx context.Context, exec executor.Executor, req *domain.Request, warnings []string, opts Options) (*Result, error) {
	// Apply URL override
	if opts.URLOverride != "" {
		req.URL = opts.URLOverride
	}

	// Execute the request
	resp, err := exec.Execute(ctx, req)
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

// RunChain executes a sequence of steps
func RunChain(ctx context.Context, factory *executor.Factory, steps []config.ChainStep, opts Options) (*ChainResult, error) {
	chainCtx := NewChainContext()
	chainResult := &ChainResult{
		Results:   make([]*Result, 0, len(steps)),
		StepNames: make([]string, 0, len(steps)),
	}

	for i, step := range steps {
		fmt.Fprintf(os.Stderr, "Running step %d: %s...\n", i+1, step.Name)

		// 1. Interpolate Strings
		finalURL, err := chainCtx.ExpandVariables(step.URL)
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		finalHeaders := make(map[string]string)
		for k, v := range step.Headers {
			expanded, err := chainCtx.ExpandVariables(v)
			if err != nil {
				return nil, fmt.Errorf("step '%s' header '%s': %w", step.Name, k, err)
			}
			finalHeaders[k] = expanded
		}

		// Interpolate body if it contains string values
		finalBody, err := interpolateBody(chainCtx, step.Body)
		if err != nil {
			return nil, fmt.Errorf("step '%s' body: %w", step.Name, err)
		}

		// Interpolate JSON string if present
		finalJSON := step.JSON
		if finalJSON != "" {
			finalJSON, err = chainCtx.ExpandVariables(finalJSON)
			if err != nil {
				return nil, fmt.Errorf("step '%s' json: %w", step.Name, err)
			}
		}

		// 2. Build request manually (don't use ToDomain as we've already interpolated)
		method := step.Method
		if method == "" {
			method = constants.MethodGET
		}

		var bodyReader io.Reader
		var contentType string

		if finalJSON != "" {
			contentType = "application/json"
			bodyReader = strings.NewReader(finalJSON)
		} else if finalBody != nil {
			contentType = "application/json"
			bodyBytes, err := json.Marshal(finalBody)
			if err != nil {
				return nil, fmt.Errorf("step '%s': invalid json in body: %w", step.Name, err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req := &domain.Request{
			URL:      finalURL,
			Method:   constants.CanonicalizeMethod(method),
			Headers:  finalHeaders,
			Body:     bodyReader,
			Metadata: make(map[string]string),
		}

		if contentType != "" {
			if req.Headers == nil {
				req.Headers = make(map[string]string)
			}
			if _, ok := req.Headers["Content-Type"]; !ok {
				req.Headers["Content-Type"] = contentType
			}
		}

		// Handle GraphQL Metadata
		if step.Graphql != "" {
			req.Metadata["transport"] = constants.TransportGraphQL
			req.Metadata["graphql_query"] = step.Graphql
			if step.Variables != nil {
				vars, _ := json.Marshal(step.Variables)
				req.Metadata["graphql_variables"] = string(vars)
			}
		} else {
			req.Metadata["transport"] = constants.TransportHTTP
		}

		// 3. Create executor for this step's transport
		exec, err := factory.Create(req.Metadata["transport"])
		if err != nil {
			return nil, fmt.Errorf("step '%s': %w", step.Name, err)
		}

		// 4. Execute
		result, err := Run(ctx, exec, req, []string{}, opts)
		if err != nil {
			return nil, fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}

		// 5. Assert Expectations
		expectRes := CheckExpectations(step.Expect, result)
		if expectRes.Error != nil {
			return nil, fmt.Errorf("step '%s' assertion failed: %w", step.Name, expectRes.Error)
		}

		// 6. Store Result
		chainCtx.AddResult(step.Name, result)
		chainResult.Results = append(chainResult.Results, result)
		chainResult.StepNames = append(chainResult.StepNames, step.Name)
	}

	return chainResult, nil
}

// interpolateBody recursively interpolates variables in body map
func interpolateBody(chainCtx *ChainContext, body map[string]interface{}) (map[string]interface{}, error) {
	if body == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for k, v := range body {
		switch val := v.(type) {
		case string:
			expanded, err := chainCtx.ExpandVariables(val)
			if err != nil {
				return nil, err
			}
			result[k] = expanded
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

// Package output provides response formatting and syntax highlighting.
package output

import (
	"encoding/json"
	"fmt"

	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/validation"
)

// AssertionResult represents a single assertion result in JSON output.
type AssertionResult struct {
	Expression    string `json:"expression"`
	Passed        bool   `json:"passed"`
	ActualValue   string `json:"actual,omitempty"`
	ExpectedValue string `json:"expected,omitempty"`
	LeftSide      string `json:"leftSide,omitempty"`
	Operator      string `json:"operator,omitempty"`
	Error         string `json:"error,omitempty"`
}

// JSONOutput represents the structured JSON output for --json flag.
type JSONOutput struct {
	Success     bool              `json:"success"`
	Body        string            `json:"body"`
	Transport   string            `json:"transport,omitempty"`
	StatusCode  int               `json:"statusCode,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	RequestURL  string            `json:"requestUrl,omitempty"`
	Method      string            `json:"method,omitempty"`
	Service     string            `json:"service,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	SizeBytes   int               `json:"sizeBytes,omitempty"`
	SizeLines   int               `json:"sizeLines,omitempty"`
	SizeChars   int               `json:"sizeChars,omitempty"`
	Timing      int64             `json:"timing"` // milliseconds
	Warnings    []string          `json:"warnings,omitempty"`
	Error       string            `json:"error,omitempty"`
	Assertions  *struct {
		Total   int               `json:"total"`
		Passed  int               `json:"passed"`
		Results []AssertionResult `json:"results,omitempty"`
	} `json:"assertions,omitempty"`
}

// JSONParams holds parameters for building JSON output.
type JSONParams struct {
	Result      *runner.Result
	ChainResult *runner.ChainResult
	ExpectRes   *runner.ExpectationResult
	Analysis    *validation.Analysis
	ExecErr     error
}

// PrintJSON outputs the result as structured JSON (handles both single and chain results).
// Returns the execution error if any occurred (even after printing successfully).
func PrintJSON(params JSONParams) error {
	// Success means we got a response - assertion failures are NOT errors
	// Only true execution failures (network errors, etc.) should set success to false
	hasResult := params.Result != nil || (params.ChainResult != nil && len(params.ChainResult.Results) > 0)
	isAssertionFailure := params.ExpectRes != nil && params.ExpectRes.Error != nil

	output := JSONOutput{
		Success: hasResult && (params.ExecErr == nil || isAssertionFailure),
	}

	// Handle chain results
	if params.ChainResult != nil && len(params.ChainResult.Results) > 0 {
		output.Transport = "chain"

		// Calculate total timing
		var totalTiming int64
		for _, r := range params.ChainResult.Results {
			totalTiming += r.Duration.Milliseconds()
		}
		output.Timing = totalTiming

		// Build combined body as JSON array of step results
		var stepBodies []any
		for i, r := range params.ChainResult.Results {
			stepBody := map[string]any{
				"step": params.ChainResult.StepNames[i],
			}
			var bodyJSON any
			if err := json.Unmarshal([]byte(r.Body), &bodyJSON); err == nil {
				stepBody["body"] = bodyJSON
			} else {
				stepBody["body"] = r.Body
			}
			stepBody["statusCode"] = r.StatusCode
			stepBody["timing"] = r.Duration.Milliseconds()
			stepBodies = append(stepBodies, stepBody)
		}
		bodyBytes, _ := json.MarshalIndent(stepBodies, "", "  ")
		output.Body = string(bodyBytes)

		// Use last step's result for metadata
		lastResult := params.ChainResult.Results[len(params.ChainResult.Results)-1]
		output.StatusCode = lastResult.StatusCode
		output.Headers = lastResult.Headers
		output.RequestURL = lastResult.RequestURL
		output.ContentType = lastResult.ContentType
		output.SizeBytes = lastResult.BodyBytes
		output.SizeLines = lastResult.BodyLines
		output.SizeChars = lastResult.BodyChars
		output.Warnings = lastResult.Warnings
	} else if params.Result != nil {
		// Handle single result
		output.Timing = params.Result.Duration.Milliseconds()
		output.Body = params.Result.Body
		output.StatusCode = params.Result.StatusCode
		output.Headers = params.Result.Headers
		output.RequestURL = params.Result.RequestURL
		output.ContentType = params.Result.ContentType
		output.SizeBytes = params.Result.BodyBytes
		output.SizeLines = params.Result.BodyLines
		output.SizeChars = params.Result.BodyChars
		output.Warnings = params.Result.Warnings

		// Determine transport type from config
		if params.Analysis != nil && params.Analysis.Base != nil {
			cfg := params.Analysis.Base
			//nolint:gocritic // ifElseChain: switch not suitable for these independent boolean conditions
			if cfg.Graphql != "" {
				output.Transport = "graphql"
			} else if cfg.Service != "" || cfg.RPC != "" {
				output.Transport = "grpc"
				output.Service = cfg.Service
			} else if cfg.Data != "" {
				output.Transport = "tcp"
			} else {
				output.Transport = "http"
			}
		}

		if params.Analysis != nil && params.Analysis.Request != nil {
			output.Method = params.Analysis.Request.Method
		}
	}

	// Add assertions if present
	if params.ExpectRes != nil {
		// Convert assertion results to JSON format
		var results []AssertionResult
		for _, ar := range params.ExpectRes.AssertionResults {
			result := AssertionResult{
				Expression:    ar.Expression,
				Passed:        ar.Passed,
				ActualValue:   ar.ActualValue,
				ExpectedValue: ar.ExpectedValue,
				LeftSide:      ar.LeftSide,
				Operator:      ar.Operator,
			}
			if ar.Error != nil {
				result.Error = ar.Error.Error()
			}
			results = append(results, result)
		}

		output.Assertions = &struct {
			Total   int               `json:"total"`
			Passed  int               `json:"passed"`
			Results []AssertionResult `json:"results,omitempty"`
		}{
			Total:   params.ExpectRes.AssertionsTotal,
			Passed:  params.ExpectRes.AssertionsPassed,
			Results: results,
		}
	}

	// Add error if execution failed (but not for assertion failures - those are shown in assertions)
	if params.ExecErr != nil && !isAssertionFailure {
		output.Error = params.ExecErr.Error()
	}

	// Marshal and print JSON
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return params.ExecErr
}

// Package filter provides JQ filtering for response bodies.
package filter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/itchyny/gojq"
)

// ApplyJQ applies a jq filter expression to the given JSON input string.
// Returns the filtered result as a string.
// If the filter produces multiple values, they are joined with newlines.
func ApplyJQ(input string, filterExpr string) (string, error) {
	return ApplyJQWithVars(input, filterExpr, nil)
}

// ApplyJQWithVars applies a jq filter expression with optional variables.
// Variables is a map of variable names to values (e.g., map[string]any{"_headers": {...}}).
func ApplyJQWithVars(input string, filterExpr string, variables map[string]any) (string, error) {
	filterExpr = strings.TrimSpace(filterExpr)
	if filterExpr == "" {
		return input, nil
	}

	// Parse the jq query
	query, err := gojq.Parse(filterExpr)
	if err != nil {
		return "", fmt.Errorf("failed to parse jq filter %q: %w", filterExpr, err)
	}

	// Compile the query with variables if provided
	if variables != nil {
		var varNames []string
		for name := range variables {
			varNames = append(varNames, "$"+name)
		}
		code, err := gojq.Compile(query, gojq.WithVariables(varNames))
		if err != nil {
			return "", fmt.Errorf("failed to compile jq filter with variables: %w", err)
		}

		// Parse the input JSON, preserving number precision
		inputData, err := parseJSONPreserveNumbers(input)
		if err != nil {
			return "", fmt.Errorf("failed to parse input as JSON: %w", err)
		}

		// Build variable values in order
		varValues := make([]any, 0, len(variables))
		for _, name := range varNames {
			varValues = append(varValues, variables[strings.TrimPrefix(name, "$")])
		}

		// Run the compiled query with variables
		iter := code.Run(inputData, varValues...)
		return collectResults(iter)
	}

	// Parse the input JSON, preserving number precision
	inputData, err := parseJSONPreserveNumbers(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input as JSON: %w", err)
	}

	// Run the query without variables
	iter := query.Run(inputData)
	return collectResults(iter)
}

// collectResults collects results from a JQ iterator
func collectResults(iter gojq.Iter) (string, error) {
	var results []string
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return "", fmt.Errorf("jq filter error: %w", err)
		}

		// Format the output
		output, err := formatOutput(v)
		if err != nil {
			return "", fmt.Errorf("failed to format jq output: %w", err)
		}
		results = append(results, output)
	}

	return strings.Join(results, "\n"), nil
}

// parseJSONPreserveNumbers parses JSON input while preserving large integer precision.
// It converts json.Number to appropriate Go types that gojq can handle.
func parseJSONPreserveNumbers(input string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()

	var data any
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	return convertNumbers(data), nil
}

// convertNumbers recursively converts json.Number to *big.Int or float64 as appropriate.
// gojq supports *big.Int for arbitrary-precision integers.
func convertNumbers(v any) any {
	switch val := v.(type) {
	case json.Number:
		// Try to parse as big.Int first for arbitrary precision
		if i, ok := new(big.Int).SetString(string(val), 10); ok {
			// Check if it fits in int (gojq prefers int for small numbers)
			if i.IsInt64() {
				return int(i.Int64())
			}
			return i
		}
		// Fall back to float64
		f, _ := val.Float64()
		return f
	case map[string]any:
		for k, v := range val {
			val[k] = convertNumbers(v)
		}
		return val
	case []any:
		for i, v := range val {
			val[i] = convertNumbers(v)
		}
		return val
	default:
		return val
	}
}

// AssertionDetail contains detailed information about an assertion failure
type AssertionDetail struct {
	Expression    string // The full assertion expression
	LeftSide      string // The left side of comparison (e.g., ".id")
	Operator      string // The operator (e.g., "==", "!=", ">", etc.)
	RightSide     string // The right side/expected value (e.g., "999")
	ActualValue   string // The actual value from the left side evaluation
	ExpectedValue string // The expected value (right side)
}

// EvalJQBoolWithDetail evaluates a JQ expression and returns detailed information about the assertion.
// This is useful for generating helpful error messages when assertions fail.
func EvalJQBoolWithDetail(input string, expr string) (bool, *AssertionDetail, error) {
	return EvalJQBoolWithDetailAndVars(input, expr, nil)
}

// parseAssertionOperator extracts the left side, operator, and right side from an assertion expression
func parseAssertionOperator(expr string, detail *AssertionDetail) {
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}
	for _, op := range operators {
		if idx := strings.Index(expr, op); idx != -1 {
			validMatch := true
			if op == "=" || op == ">" || op == "<" {
				if idx > 0 && (expr[idx-1] == '>' || expr[idx-1] == '<' || expr[idx-1] == '!' || expr[idx-1] == '=') {
					validMatch = false
				}
				if idx < len(expr)-1 && expr[idx+1] == '=' {
					validMatch = false
				}
			}
			if validMatch {
				detail.LeftSide = strings.TrimSpace(expr[:idx])
				detail.Operator = op
				detail.RightSide = strings.TrimSpace(expr[idx+len(op):])
				detail.ExpectedValue = detail.RightSide
				break
			}
		}
	}
}

// evalLeftSide evaluates the left side of an assertion to get the actual value
func evalLeftSide(leftSide string, inputData any, varNames []string, varValues []any) string {
	if leftSide == "" {
		return ""
	}
	leftQuery, err := gojq.Parse(leftSide)
	if err != nil {
		return ""
	}

	var leftIter gojq.Iter
	if varNames != nil {
		leftCode, err := gojq.Compile(leftQuery, gojq.WithVariables(varNames))
		if err != nil {
			return ""
		}
		leftIter = leftCode.Run(inputData, varValues...)
	} else {
		leftIter = leftQuery.Run(inputData)
	}

	if leftVal, ok := leftIter.Next(); ok {
		if _, isErr := leftVal.(error); !isErr {
			return formatValue(leftVal)
		}
	}
	return ""
}

// compileAndRunWithVars compiles a query with variables and returns the iterator
func compileAndRunWithVars(query *gojq.Query, inputData any, variables map[string]any) (gojq.Iter, []string, []any, error) {
	var varNames []string
	for name := range variables {
		varNames = append(varNames, "$"+name)
	}
	code, err := gojq.Compile(query, gojq.WithVariables(varNames))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to compile jq expression with variables: %w", err)
	}

	varValues := make([]any, 0, len(variables))
	for _, name := range varNames {
		varValues = append(varValues, variables[strings.TrimPrefix(name, "$")])
	}

	return code.Run(inputData, varValues...), varNames, varValues, nil
}

// EvalJQBoolWithDetailAndVars evaluates a JQ expression with optional variables.
// Variables is a map of variable names to values (e.g., map[string]any{"_headers": {...}}).
func EvalJQBoolWithDetailAndVars(input string, expr string, variables map[string]any) (bool, *AssertionDetail, error) {
	expr = strings.TrimSpace(expr)
	detail := &AssertionDetail{Expression: expr}

	if expr == "" {
		return false, detail, fmt.Errorf("empty assertion expression")
	}

	parseAssertionOperator(expr, detail)

	query, err := gojq.Parse(expr)
	if err != nil {
		return false, detail, fmt.Errorf("failed to parse jq expression %q: %w", expr, err)
	}

	inputData, err := parseJSONPreserveNumbers(input)
	if err != nil {
		return false, detail, fmt.Errorf("failed to parse input as JSON: %w", err)
	}

	var iter gojq.Iter
	var varNames []string
	var varValues []any

	if variables != nil {
		iter, varNames, varValues, err = compileAndRunWithVars(query, inputData, variables)
		if err != nil {
			return false, detail, err
		}
		detail.ActualValue = evalLeftSide(detail.LeftSide, inputData, varNames, varValues)
	} else {
		iter = query.Run(inputData)
		detail.ActualValue = evalLeftSide(detail.LeftSide, inputData, nil, nil)
	}

	v, ok := iter.Next()
	if !ok {
		return false, detail, fmt.Errorf("assertion %q produced no result", expr)
	}
	if err, isErr := v.(error); isErr {
		return false, detail, fmt.Errorf("assertion error: %w", err)
	}

	if val, ok := v.(bool); ok {
		return val, detail, nil
	}
	return false, detail, fmt.Errorf("assertion %q did not return boolean (got %T: %v)", expr, v, v)
}

// formatValue formats a value for display in error messages
func formatValue(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case int, int64, float64:
		return fmt.Sprintf("%v", val)
	case *big.Int:
		return val.String()
	default:
		// For complex types, use JSON encoding
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

// formatOutput converts a value to its JSON string representation.
// Strings are returned without quotes for cleaner output.
func formatOutput(v any) (string, error) {
	if v == nil {
		return "null", nil
	}

	switch val := v.(type) {
	case string:
		// Return strings without quotes for cleaner output
		return val, nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case float64:
		// Use %v for cleaner output (no trailing zeros for whole numbers)
		return fmt.Sprintf("%v", val), nil
	case *big.Int:
		return val.String(), nil
	default:
		// For complex types (objects, arrays), use JSON encoding
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

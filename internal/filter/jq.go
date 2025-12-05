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
	filterExpr = strings.TrimSpace(filterExpr)
	if filterExpr == "" {
		return input, nil
	}

	// Parse the jq query
	query, err := gojq.Parse(filterExpr)
	if err != nil {
		return "", fmt.Errorf("failed to parse jq filter %q: %w", filterExpr, err)
	}

	// Parse the input JSON, preserving number precision
	inputData, err := parseJSONPreserveNumbers(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input as JSON: %w", err)
	}

	// Run the query
	iter := query.Run(inputData)

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

// EvalJQBool evaluates a JQ expression and returns true if it evaluates to boolean true.
// Used for assertion checking in chains.
func EvalJQBool(input string, expr string) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, fmt.Errorf("empty assertion expression")
	}

	// Parse the jq query
	query, err := gojq.Parse(expr)
	if err != nil {
		return false, fmt.Errorf("failed to parse jq expression %q: %w", expr, err)
	}

	// Parse the input JSON
	inputData, err := parseJSONPreserveNumbers(input)
	if err != nil {
		return false, fmt.Errorf("failed to parse input as JSON: %w", err)
	}

	// Run the query
	iter := query.Run(inputData)
	v, ok := iter.Next()
	if !ok {
		return false, fmt.Errorf("assertion %q produced no result", expr)
	}
	if err, isErr := v.(error); isErr {
		return false, fmt.Errorf("assertion error: %w", err)
	}

	// Check if result is boolean true
	switch val := v.(type) {
	case bool:
		return val, nil
	default:
		return false, fmt.Errorf("assertion %q did not return boolean (got %T: %v)", expr, v, v)
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

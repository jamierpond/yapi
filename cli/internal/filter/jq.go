package filter

import (
	"encoding/json"
	"fmt"
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

	// Parse the input JSON
	var inputData any
	if err := json.Unmarshal([]byte(input), &inputData); err != nil {
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
	case bool, int, int64, float64:
		return fmt.Sprintf("%v", val), nil
	default:
		// For complex types (objects, arrays), use JSON encoding
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

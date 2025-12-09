package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"yapi.run/cli/internal/vars"
)

type StepResult struct {
	BodyRaw    string
	BodyJSON   map[string]any
	Headers    map[string]string
	StatusCode int
}

type ChainContext struct {
	Results map[string]StepResult
}

func NewChainContext() *ChainContext {
	return &ChainContext{
		Results: make(map[string]StepResult),
	}
}

func (c *ChainContext) AddResult(name string, result *Result) {
	sr := StepResult{
		BodyRaw:    result.Body,
		Headers:    make(map[string]string),
		StatusCode: result.StatusCode,
	}

	// Copy all response headers
	for k, v := range result.Headers {
		sr.Headers[k] = v
	}

	var data map[string]any
	// Try parsing JSON; ignore errors (BodyJSON stays nil)
	if err := json.Unmarshal([]byte(result.Body), &data); err == nil {
		sr.BodyJSON = data
	}
	c.Results[name] = sr
}

// ExpandVariables replaces $var and ${var} with values from Env or Chain Context.
func (c *ChainContext) ExpandVariables(input string) (string, error) {
	var capturedErr error

	result := vars.Expansion.ReplaceAllStringFunc(input, func(match string) string {
		var key string
		if strings.HasPrefix(match, "${") {
			// Strict: ${key}
			key = match[2 : len(match)-1]
		} else {
			// Lazy: $key
			key = match[1:]
		}

		// 1. Check OS Environment
		if val, ok := os.LookupEnv(key); ok {
			return val
		}

		// 2. Check Chain Context (must contain dot)
		if strings.Contains(key, ".") {
			val, err := c.resolveChainVar(key)
			if err != nil {
				if capturedErr == nil {
					capturedErr = err
				}
				return match // Return original on error
			}
			return val
		}

		// Not found: return as is (or could error if strict mode)
		return match
	})

	if capturedErr != nil {
		return "", capturedErr
	}
	return result, nil
}

func (c *ChainContext) resolveChainVar(key string) (string, error) {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid reference format '%s'", key)
	}

	stepName := parts[0]
	path := parts[1:]

	res, ok := c.Results[stepName]
	if !ok {
		return "", fmt.Errorf("step '%s' not found (or hasn't run yet)", stepName)
	}

	// 1. Reserved Keywords
	if len(path) == 1 {
		switch path[0] {
		case "body":
			return res.BodyRaw, nil
		case "status":
			return strconv.Itoa(res.StatusCode), nil
		}
	}

	// 2. Headers - check HTTP response headers first, then fall back to JSON body
	if path[0] == "headers" {
		if len(path) < 2 {
			return "", fmt.Errorf("header reference requires key (e.g. headers.Content-Type)")
		}
		target := path[1]
		// Try exact match in HTTP response headers
		if v, ok := res.Headers[target]; ok {
			return v, nil
		}
		// Try case-insensitive in HTTP response headers
		for k, v := range res.Headers {
			if strings.EqualFold(k, target) {
				return v, nil
			}
		}
		// Fall back to JSON path lookup (for APIs like httpbin that echo headers in body)
		if res.BodyJSON != nil {
			val, err := jsonPathLookup(res.BodyJSON, path)
			if err == nil {
				return val, nil
			}
		}
		return "", fmt.Errorf("header '%s' not found in step '%s'", target, stepName)
	}

	// 3. JSON Path
	if res.BodyJSON == nil {
		return "", fmt.Errorf("step '%s' did not return JSON, cannot access property '%s'", stepName, key)
	}

	return jsonPathLookup(res.BodyJSON, path)
}

func jsonPathLookup(data any, path []string) (string, error) {
	current := data
	for i, key := range path {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[key]
			if !ok {
				return "", fmt.Errorf("key '%s' not found at path '%s'", key, strings.Join(path[:i+1], "."))
			}
			current = val
		default:
			return "", fmt.Errorf("path segment '%s' is not an object", strings.Join(path[:i], "."))
		}
	}
	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v, nil
	case float64:
		// Check if it's actually an integer
		if v == float64(int(v)) {
			return strconv.Itoa(int(v)), nil
		}
		return fmt.Sprintf("%v", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	case nil:
		return "null", nil
	default:
		// For complex types, marshal to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v), nil
		}
		return string(jsonBytes), nil
	}
}

// ResolveVariableRaw checks if input is a pure variable reference (e.g. "$step.field" or "${step.field}")
// and returns the raw typed value. Returns (value, true) if resolved, (nil, false) otherwise.
func (c *ChainContext) ResolveVariableRaw(input string) (any, bool) {
	trimmed := strings.TrimSpace(input)

	// Check if it's a pure reference (entire string is just the variable)
	match := vars.Expansion.FindStringSubmatch(trimmed)
	if match == nil {
		return nil, false
	}

	// Verify the entire string is just the variable reference
	if match[0] != trimmed {
		return nil, false
	}

	var key string
	if match[1] != "" {
		// Strict format: ${key}
		key = match[1]
	} else {
		// Lazy format: $key
		key = match[2]
	}

	// Must contain a dot to be a chain reference
	if !strings.Contains(key, ".") {
		return nil, false
	}

	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return nil, false
	}

	stepName := parts[0]
	path := parts[1:]

	res, ok := c.Results[stepName]
	if !ok {
		return nil, false
	}

	// JSON Path lookup returning raw value
	if res.BodyJSON == nil {
		return nil, false
	}

	val, err := jsonPathLookupRaw(res.BodyJSON, path)
	if err != nil {
		return nil, false
	}

	return val, true
}

// jsonPathLookupRaw returns the raw typed value at the given path
func jsonPathLookupRaw(data any, path []string) (any, error) {
	current := data
	for i, key := range path {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[key]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found at path '%s'", key, strings.Join(path[:i+1], "."))
			}
			current = val
		default:
			return nil, fmt.Errorf("path segment '%s' is not an object", strings.Join(path[:i], "."))
		}
	}
	return current, nil
}

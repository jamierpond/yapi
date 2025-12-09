package runner

import (
	"os"
	"testing"
)

func TestChainContext_ExpandVariables_EnvVars(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	ctx := NewChainContext()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple env var with $",
			input:    "$TEST_VAR",
			expected: "test_value",
		},
		{
			name:     "env var with ${...}",
			input:    "${TEST_VAR}",
			expected: "test_value",
		},
		{
			name:     "env var in URL",
			input:    "https://example.com/$TEST_VAR/path",
			expected: "https://example.com/test_value/path",
		},
		{
			name:     "unknown env var stays as is",
			input:    "$UNKNOWN_VAR",
			expected: "$UNKNOWN_VAR",
		},
		{
			name:     "no variables",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ctx.ExpandVariables(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ExpandVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChainContext_ExpandVariables_ChainRefs(t *testing.T) {
	ctx := NewChainContext()

	// Add a step result
	ctx.Results["login"] = StepResult{
		BodyRaw:    `{"access_token":"abc123","user":{"id":42,"name":"test"}}`,
		BodyJSON:   map[string]any{"access_token": "abc123", "user": map[string]any{"id": float64(42), "name": "test"}},
		Headers:    map[string]string{"Content-Type": "application/json", "X-Custom": "custom-value"},
		StatusCode: 200,
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "access JSON property",
			input:    "${login.access_token}",
			expected: "abc123",
		},
		{
			name:     "access nested JSON property",
			input:    "${login.user.name}",
			expected: "test",
		},
		{
			name:     "access numeric JSON property",
			input:    "${login.user.id}",
			expected: "42",
		},
		{
			name:     "access body raw",
			input:    "${login.body}",
			expected: `{"access_token":"abc123","user":{"id":42,"name":"test"}}`,
		},
		{
			name:     "access status",
			input:    "${login.status}",
			expected: "200",
		},
		{
			name:     "access header",
			input:    "${login.headers.Content-Type}",
			expected: "application/json",
		},
		{
			name:     "access custom header",
			input:    "${login.headers.X-Custom}",
			expected: "custom-value",
		},
		{
			name:     "bearer token header",
			input:    "Bearer ${login.access_token}",
			expected: "Bearer abc123",
		},
		{
			name:    "reference undefined step",
			input:   "${undefined.token}",
			wantErr: true,
		},
		{
			name:    "reference non-existent JSON key",
			input:   "${login.nonexistent}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ctx.ExpandVariables(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ExpandVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChainContext_AddResult(t *testing.T) {
	ctx := NewChainContext()

	result := &Result{
		Body:        `{"message":"success"}`,
		ContentType: "application/json",
		StatusCode:  200,
	}

	ctx.AddResult("step1", result)

	stored, ok := ctx.Results["step1"]
	if !ok {
		t.Fatal("Result not stored")
	}

	if stored.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", stored.StatusCode)
	}

	if stored.BodyRaw != `{"message":"success"}` {
		t.Errorf("BodyRaw = %s, want {\"message\":\"success\"}", stored.BodyRaw)
	}

	if stored.BodyJSON == nil {
		t.Error("BodyJSON should not be nil for valid JSON")
	}

	if stored.BodyJSON["message"] != "success" {
		t.Errorf("BodyJSON[message] = %v, want success", stored.BodyJSON["message"])
	}
}

func TestChainContext_AddResult_NonJSON(t *testing.T) {
	ctx := NewChainContext()

	result := &Result{
		Body:        "plain text response",
		ContentType: "text/plain",
		StatusCode:  200,
	}

	ctx.AddResult("step1", result)

	stored := ctx.Results["step1"]
	if stored.BodyJSON != nil {
		t.Error("BodyJSON should be nil for non-JSON response")
	}
}

func TestJsonPathLookup(t *testing.T) {
	data := map[string]any{
		"string":  "value",
		"number":  float64(42),
		"boolean": true,
		"nested": map[string]any{
			"deep": "nested_value",
		},
	}

	tests := []struct {
		name     string
		path     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "string value",
			path:     []string{"string"},
			expected: "value",
		},
		{
			name:     "number value",
			path:     []string{"number"},
			expected: "42",
		},
		{
			name:     "boolean value",
			path:     []string{"boolean"},
			expected: "true",
		},
		{
			name:     "nested value",
			path:     []string{"nested", "deep"},
			expected: "nested_value",
		},
		{
			name:    "non-existent key",
			path:    []string{"nonexistent"},
			wantErr: true,
		},
		{
			name:    "non-existent nested key",
			path:    []string{"nested", "nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonPathLookup(data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonPathLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("jsonPathLookup() = %v, want %v", result, tt.expected)
			}
		})
	}
}

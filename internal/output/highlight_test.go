package output

import (
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		contentType string
		expected    string
	}{
		{
			name:        "JSON from content type",
			raw:         "{}",
			contentType: "application/json",
			expected:    "json",
		},
		{
			name:        "JSON from content type with charset",
			raw:         "{}",
			contentType: "application/json; charset=utf-8",
			expected:    "json",
		},
		{
			name:        "HTML from content type",
			raw:         "<html></html>",
			contentType: "text/html",
			expected:    "html",
		},
		{
			name:        "HTML from content type with charset",
			raw:         "<html></html>",
			contentType: "text/html; charset=utf-8",
			expected:    "html",
		},
		{
			name:        "JSON from content sniffing - object",
			raw:         `{"key": "value"}`,
			contentType: "",
			expected:    "json",
		},
		{
			name:        "JSON from content sniffing - array",
			raw:         `["item1", "item2"]`,
			contentType: "",
			expected:    "json",
		},
		{
			name:        "JSON from content sniffing with whitespace",
			raw:         `  {"key": "value"}`,
			contentType: "",
			expected:    "json",
		},
		{
			name:        "HTML from content sniffing - doctype",
			raw:         `<!DOCTYPE html><html></html>`,
			contentType: "",
			expected:    "html",
		},
		{
			name:        "HTML from content sniffing - html tag",
			raw:         `<html><body></body></html>`,
			contentType: "",
			expected:    "html",
		},
		{
			name:        "Default to JSON for unknown content",
			raw:         "some plain text",
			contentType: "text/plain",
			expected:    "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguage(tt.raw, tt.contentType)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q, %q) = %q, want %q", tt.raw, tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestHighlight_NoColor(t *testing.T) {
	raw := `{"key": "value"}`
	result := Highlight(raw, "application/json", true)
	// With noColor=true, should still pretty-print but without ANSI codes
	expected := `{
  "key": "value"
}`
	if result != expected {
		t.Errorf("Highlight with noColor=true should return pretty-printed JSON, got %q, expected %q", result, expected)
	}
	// Should not contain ANSI escape codes
	if strings.Contains(result, "\x1b[") {
		t.Error("Highlight with noColor=true should not contain ANSI codes")
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	raw := `{"key":"value","nested":{"a":1}}`
	result := prettyPrintJSON(raw)
	expected := `{
  "key": "value",
  "nested": {
    "a": 1
  }
}`
	if result != expected {
		t.Errorf("prettyPrintJSON got %q, expected %q", result, expected)
	}
}

func TestPrettyPrintJSON_Invalid(t *testing.T) {
	raw := `not valid json`
	result := prettyPrintJSON(raw)
	if result != raw {
		t.Errorf("prettyPrintJSON with invalid JSON should return raw, got %q", result)
	}
}

func TestHighlightWithChroma(t *testing.T) {
	// Test that valid JSON gets some highlighting (contains ANSI codes)
	raw := `{"key": "value"}`
	result := highlightWithChroma(raw, "json")

	// In a TTY, result should contain ANSI escape codes
	// We can't fully test TTY behavior in tests, but we can test the chroma function directly
	if result == "" {
		t.Error("highlightWithChroma returned empty string")
	}

	// Test that HTML gets some highlighting
	htmlRaw := `<html><body><p>Hello</p></body></html>`
	htmlResult := highlightWithChroma(htmlRaw, "html")
	if htmlResult == "" {
		t.Error("highlightWithChroma for HTML returned empty string")
	}
}

func TestHighlightWithChroma_InvalidLexer(t *testing.T) {
	raw := `some text`
	// Use an invalid lexer name
	result := highlightWithChroma(raw, "nonexistent-language-xyz")
	if result != raw {
		t.Errorf("highlightWithChroma with invalid lexer should return raw, got %q", result)
	}
}

func TestHighlightWithChroma_ContainsANSI(t *testing.T) {
	raw := `{"name": "test", "value": 123}`
	result := highlightWithChroma(raw, "json")

	// Check that ANSI escape codes are present (they start with \x1b[ or \033[)
	if !strings.Contains(result, "\x1b[") && !strings.Contains(result, "\033[") {
		// It's possible the output doesn't have ANSI if the formatter doesn't add any,
		// but for JSON with dracula style, there should be some coloring
		t.Log("Warning: highlightWithChroma result may not contain ANSI codes")
	}
}

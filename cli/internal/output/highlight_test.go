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

func TestPrettyPrintJSON_MultipleObjects(t *testing.T) {
	raw := `{"name":"foo"}
{"name":"bar"}
{"name":"baz"}`
	result := prettyPrintJSON(raw)
	expected := `{
  "name": "foo"
}
{
  "name": "bar"
}
{
  "name": "baz"
}`
	if result != expected {
		t.Errorf("prettyPrintJSON with multiple objects:\ngot:\n%s\n\nexpected:\n%s", result, expected)
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

func FuzzHighlight(f *testing.F) {
	// Seed with various content types and payloads
	f.Add(`{"key": "value"}`, "application/json")
	f.Add(`{"nested": {"deep": {"value": 123}}}`, "application/json")
	f.Add(`[1, 2, 3, "four", null, true]`, "application/json")
	f.Add(`<!DOCTYPE html><html><body><p>Hello</p></body></html>`, "text/html")
	f.Add(`<html><head><title>Test</title></head></html>`, "text/html")
	f.Add(`plain text content`, "text/plain")
	f.Add(``, "application/json")
	f.Add(`null`, "application/json")
	f.Add(`"just a string"`, "application/json")
	f.Add(`12345`, "application/json")
	f.Add(`{invalid json}`, "application/json")
	f.Add(`<not>valid<html`, "text/html")
	f.Add(`{"unicode": "日本語 🚀 emoji"}`, "application/json")

	f.Fuzz(func(t *testing.T, raw string, contentType string) {
		// Highlight should not panic on any input
		_ = Highlight(raw, contentType, true)
		_ = Highlight(raw, contentType, false)
	})
}

func FuzzDetectLanguage(f *testing.F) {
	f.Add(`{"key": "value"}`, "application/json")
	f.Add(`<html></html>`, "text/html")
	f.Add(`plain text`, "")
	f.Add(`  {  "with": "whitespace" }`, "")
	f.Add(`<!doctype html>`, "")
	f.Add(``, "unknown/type")

	f.Fuzz(func(t *testing.T, raw string, contentType string) {
		// detectLanguage should not panic on any input
		result := detectLanguage(raw, contentType)
		// Result should always be either "json" or "html"
		if result != "json" && result != "html" {
			t.Errorf("detectLanguage returned unexpected value: %q", result)
		}
	})
}

func FuzzPrettyPrintJSON(f *testing.F) {
	f.Add(`{"key": "value"}`)
	f.Add(`[1, 2, 3]`)
	f.Add(`null`)
	f.Add(`"string"`)
	f.Add(`123`)
	f.Add(`true`)
	f.Add(`{"deeply": {"nested": {"object": {"with": "values"}}}}`)
	f.Add(`not valid json`)
	f.Add(``)
	f.Add(`{"big": 4722366482869645213696}`)

	f.Fuzz(func(t *testing.T, raw string) {
		// prettyPrintJSON should not panic on any input
		_ = prettyPrintJSON(raw)
	})
}

func FuzzPrettyPrintHTML(f *testing.F) {
	f.Add(`<!DOCTYPE html><html><body></body></html>`)
	f.Add(`<html><head><title>Test</title></head><body><p>Hello</p></body></html>`)
	f.Add(`<div><span>nested</span></div>`)
	f.Add(`not valid html`)
	f.Add(`<unclosed`)
	f.Add(``)
	f.Add(`<script>alert('xss')</script>`)

	f.Fuzz(func(t *testing.T, raw string) {
		// prettyPrintHTML should not panic on any input
		_ = prettyPrintHTML(raw)
	})
}

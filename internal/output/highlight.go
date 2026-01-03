// Package output provides response formatting and syntax highlighting.
package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"codeberg.org/derat/htmlpretty"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"golang.org/x/net/html"
)

// Language constants for syntax highlighting
const (
	langJSON = "json"
	langHTML = "html"
)

// Highlight applies syntax highlighting and pretty-printing to the given raw string based on content type.
// If noColor is true or stdout is not a TTY, it returns pretty-printed output without colors.
func Highlight(raw string, contentType string, noColor bool) string {
	lang := detectLanguage(raw, contentType)

	// Always pretty-print, regardless of color setting
	formatted := prettyPrint(raw, lang)

	if noColor {
		return formatted
	}

	if !isTerminal() {
		return formatted
	}

	return highlightWithChroma(formatted, lang)
}

// isTerminal checks if stdout is a TTY (terminal).
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// detectLanguage determines the syntax highlighting language based on content type and content.
func detectLanguage(raw string, contentType string) string {
	// Check content type header
	if strings.Contains(contentType, "application/json") {
		return langJSON
	}
	if strings.Contains(contentType, "text/html") {
		return langHTML
	}

	// Fallback to content sniffing
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return langJSON
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return langHTML
	}

	// Default to JSON
	return langJSON
}

// highlightWithChroma applies Chroma syntax highlighting to the raw string.
func highlightWithChroma(raw string, lang string) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		return raw
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.TTY8
	if formatter == nil {
		return raw
	}

	iterator, err := lexer.Tokenise(nil, raw)
	if err != nil {
		return raw
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return raw
	}

	return buf.String()
}

// prettyPrint formats JSON and HTML content for better readability.
func prettyPrint(raw string, lang string) string {
	switch lang {
	case langJSON:
		return prettyPrintJSON(raw)
	case langHTML:
		return prettyPrintHTML(raw)
	default:
		return raw
	}
}

// prettyPrintJSON formats JSON with indentation.
// Handles multiple JSON objects in a stream (common jq output).
func prettyPrintJSON(raw string) string {
	dec := json.NewDecoder(strings.NewReader(raw))
	var results []string

	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			break
		}

		pretty, _ := json.MarshalIndent(v, "", "  ")
		results = append(results, string(pretty))
	}

	if len(results) > 0 {
		return strings.Join(results, "\n")
	}

	// Fall back to raw if nothing parsed
	return raw
}

// prettyPrintHTML formats HTML using htmlpretty.
func prettyPrintHTML(raw string) string {
	node, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return raw
	}

	var buf bytes.Buffer
	if err := htmlpretty.Print(&buf, node, "  ", 120); err != nil {
		return raw
	}

	return buf.String()
}

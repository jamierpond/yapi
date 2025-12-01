package output

import (
	"bytes"
	"os"
	"strings"

	"yapi.run/internal/filter"

	"codeberg.org/derat/htmlpretty"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"golang.org/x/net/html"
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
		return "json"
	}
	if strings.Contains(contentType, "text/html") {
		return "html"
	}

	// Fallback to content sniffing
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return "json"
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return "html"
	}

	// Default to JSON
	return "json"
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
	case "json":
		return prettyPrintJSON(raw)
	case "html":
		return prettyPrintHTML(raw)
	default:
		return raw
	}
}

// prettyPrintJSON formats JSON with indentation using jq's identity filter.
func prettyPrintJSON(raw string) string {
	// Use jq with identity filter "." to pretty-print JSON
	result, err := filter.ApplyJQ(raw, ".")
	if err != nil {
		// If it's not valid JSON, return as-is
		return raw
	}
	return result
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

package validation

import (
	"fmt"
	"io"
	"strings"

	"yapi.run/cli/internal/cli/color"
)

// ValidationError provides specific information about validation failures.
type ValidationError struct {
	Diagnostics []Diagnostic
}

func (e *ValidationError) Error() string {
	var errMsgs []string
	for _, d := range e.Diagnostics {
		if d.Severity == SeverityError {
			errMsgs = append(errMsgs, d.Message)
		}
	}
	if len(errMsgs) == 0 {
		return "validation failed"
	}
	if len(errMsgs) == 1 {
		return errMsgs[0]
	}
	return fmt.Sprintf("%d validation errors: %s", len(errMsgs), strings.Join(errMsgs, "; "))
}

// FormatDiagnostic formats a single diagnostic with color support.
func FormatDiagnostic(d Diagnostic, noColor bool) string {
	lineInfo := ""
	if d.Line >= 0 {
		lineInfo = fmt.Sprintf(" (line %d)", d.Line+1)
	}

	var formatted string
	switch d.Severity {
	case SeverityError:
		formatted = "[ERROR]" + lineInfo + " " + d.Message
		if !noColor {
			formatted = color.Red(formatted)
		}
	case SeverityWarning:
		formatted = "[WARN]" + lineInfo + " " + d.Message
		if !noColor {
			formatted = color.Yellow(formatted)
		}
	default:
		formatted = "[INFO]" + lineInfo + " " + d.Message
		if !noColor {
			formatted = color.Cyan(formatted)
		}
	}
	return formatted
}

// PrintDiagnostics prints diagnostics filtered by a predicate.
func PrintDiagnostics(analysis *Analysis, out io.Writer, filter func(Diagnostic) bool, noColor bool) {
	if analysis == nil {
		return
	}

	for _, d := range analysis.Diagnostics {
		if filter != nil && !filter(d) {
			continue
		}
		_, _ = fmt.Fprintln(out, FormatDiagnostic(d, noColor))
	}
}

// PrintErrors prints error-level diagnostics.
func PrintErrors(analysis *Analysis, out io.Writer, noColor bool) {
	PrintDiagnostics(analysis, out, func(d Diagnostic) bool {
		return d.Severity == SeverityError
	}, noColor)
}

// PrintWarnings prints warnings and non-error diagnostics.
func PrintWarnings(analysis *Analysis, out io.Writer, noColor bool) {
	if analysis == nil {
		return
	}

	// Print legacy warnings (from parser level)
	for _, w := range analysis.Warnings {
		msg := "[WARN] " + w
		if !noColor {
			msg = color.Yellow(msg)
		}
		_, _ = fmt.Fprintln(out, msg)
	}

	// Print non-error diagnostics
	for _, d := range analysis.Diagnostics {
		if d.Severity != SeverityError {
			_, _ = fmt.Fprintln(out, FormatDiagnostic(d, noColor))
		}
	}
}

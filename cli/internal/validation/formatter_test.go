package validation

import (
	"bytes"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name        string
		diagnostics []Diagnostic
		want        string
	}{
		{
			name:        "no diagnostics",
			diagnostics: []Diagnostic{},
			want:        "validation failed",
		},
		{
			name: "single error",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "field is required"},
			},
			want: "field is required",
		},
		{
			name: "multiple errors",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "field is required"},
				{Severity: SeverityError, Message: "invalid format"},
			},
			want: "2 validation errors: field is required; invalid format",
		},
		{
			name: "mixed severity - only errors counted",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "field is required"},
				{Severity: SeverityWarning, Message: "deprecated field"},
				{Severity: SeverityError, Message: "invalid format"},
			},
			want: "2 validation errors: field is required; invalid format",
		},
		{
			name: "no errors but has warnings",
			diagnostics: []Diagnostic{
				{Severity: SeverityWarning, Message: "deprecated field"},
			},
			want: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{Diagnostics: tt.diagnostics}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDiagnostic(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic Diagnostic
		noColor    bool
		wantSubstr []string // Substrings that should be in output
	}{
		{
			name: "error with line number",
			diagnostic: Diagnostic{
				Severity: SeverityError,
				Message:  "field is required",
				Line:     5,
			},
			noColor:    true,
			wantSubstr: []string{"[ERROR]", "(line 6)", "field is required"},
		},
		{
			name: "warning without line number",
			diagnostic: Diagnostic{
				Severity: SeverityWarning,
				Message:  "deprecated field",
				Line:     -1,
			},
			noColor:    true,
			wantSubstr: []string{"[WARN]", "deprecated field"},
		},
		{
			name: "info diagnostic",
			diagnostic: Diagnostic{
				Severity: SeverityInfo,
				Message:  "suggestion to improve",
				Line:     10,
			},
			noColor:    true,
			wantSubstr: []string{"[INFO]", "(line 11)", "suggestion to improve"},
		},
		{
			name: "error at line 0",
			diagnostic: Diagnostic{
				Severity: SeverityError,
				Message:  "syntax error",
				Line:     0,
			},
			noColor:    true,
			wantSubstr: []string{"[ERROR]", "(line 1)", "syntax error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDiagnostic(tt.diagnostic, tt.noColor)
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatDiagnostic() = %v, want to contain %v", got, substr)
				}
			}
		})
	}
}

func TestFormatDiagnostic_Colors(t *testing.T) {
	diagnostic := Diagnostic{
		Severity: SeverityError,
		Message:  "test error",
		Line:     0,
	}

	withColor := FormatDiagnostic(diagnostic, false)
	withoutColor := FormatDiagnostic(diagnostic, true)

	// With color should have ANSI escape codes
	if !strings.Contains(withColor, "\x1b[") {
		t.Error("FormatDiagnostic(noColor=false) should contain ANSI color codes")
	}

	// Without color should not have ANSI escape codes
	if strings.Contains(withoutColor, "\x1b[") {
		t.Error("FormatDiagnostic(noColor=true) should not contain ANSI color codes")
	}
}

func TestPrintDiagnostics(t *testing.T) {
	analysis := &Analysis{
		Diagnostics: []Diagnostic{
			{Severity: SeverityError, Message: "error 1", Line: 0},
			{Severity: SeverityWarning, Message: "warning 1", Line: 1},
			{Severity: SeverityError, Message: "error 2", Line: 2},
		},
	}

	tests := []struct {
		name       string
		filter     func(Diagnostic) bool
		wantCount  int
		wantSubstr []string
	}{
		{
			name:       "no filter - all diagnostics",
			filter:     nil,
			wantCount:  3,
			wantSubstr: []string{"error 1", "warning 1", "error 2"},
		},
		{
			name: "only errors",
			filter: func(d Diagnostic) bool {
				return d.Severity == SeverityError
			},
			wantCount:  2,
			wantSubstr: []string{"error 1", "error 2"},
		},
		{
			name: "only warnings",
			filter: func(d Diagnostic) bool {
				return d.Severity == SeverityWarning
			},
			wantCount:  1,
			wantSubstr: []string{"warning 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			PrintDiagnostics(analysis, buf, tt.filter, true)

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")

			// Check line count
			if len(lines) != tt.wantCount {
				t.Errorf("PrintDiagnostics() output %d lines, want %d", len(lines), tt.wantCount)
			}

			// Check expected substrings
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(output, substr) {
					t.Errorf("PrintDiagnostics() output = %v, want to contain %v", output, substr)
				}
			}
		})
	}
}

func TestPrintDiagnostics_NilAnalysis(t *testing.T) {
	buf := &bytes.Buffer{}
	PrintDiagnostics(nil, buf, nil, true)

	if buf.Len() != 0 {
		t.Errorf("PrintDiagnostics(nil) should not write anything, got: %v", buf.String())
	}
}

func TestPrintErrors(t *testing.T) {
	analysis := &Analysis{
		Diagnostics: []Diagnostic{
			{Severity: SeverityError, Message: "error 1"},
			{Severity: SeverityWarning, Message: "warning 1"},
			{Severity: SeverityError, Message: "error 2"},
		},
	}

	buf := &bytes.Buffer{}
	PrintErrors(analysis, buf, true)

	output := buf.String()

	// Should contain errors
	if !strings.Contains(output, "error 1") {
		t.Error("PrintErrors() should contain 'error 1'")
	}
	if !strings.Contains(output, "error 2") {
		t.Error("PrintErrors() should contain 'error 2'")
	}

	// Should not contain warnings
	if strings.Contains(output, "warning 1") {
		t.Error("PrintErrors() should not contain warnings")
	}
}

func TestPrintWarnings(t *testing.T) {
	analysis := &Analysis{
		Warnings: []string{"legacy warning 1", "legacy warning 2"},
		Diagnostics: []Diagnostic{
			{Severity: SeverityError, Message: "error 1"},
			{Severity: SeverityWarning, Message: "diagnostic warning"},
			{Severity: SeverityInfo, Message: "info message"},
		},
	}

	buf := &bytes.Buffer{}
	PrintWarnings(analysis, buf, true)

	output := buf.String()

	// Should contain legacy warnings
	if !strings.Contains(output, "legacy warning 1") {
		t.Error("PrintWarnings() should contain 'legacy warning 1'")
	}
	if !strings.Contains(output, "legacy warning 2") {
		t.Error("PrintWarnings() should contain 'legacy warning 2'")
	}

	// Should contain non-error diagnostics
	if !strings.Contains(output, "diagnostic warning") {
		t.Error("PrintWarnings() should contain 'diagnostic warning'")
	}
	if !strings.Contains(output, "info message") {
		t.Error("PrintWarnings() should contain 'info message'")
	}

	// Should not contain errors
	if strings.Contains(output, "error 1") {
		t.Error("PrintWarnings() should not contain errors")
	}
}

func TestPrintWarnings_NilAnalysis(t *testing.T) {
	buf := &bytes.Buffer{}
	PrintWarnings(nil, buf, true)

	if buf.Len() != 0 {
		t.Errorf("PrintWarnings(nil) should not write anything, got: %v", buf.String())
	}
}

func TestPrintWarnings_EmptyWarnings(t *testing.T) {
	analysis := &Analysis{
		Warnings:    []string{},
		Diagnostics: []Diagnostic{},
	}

	buf := &bytes.Buffer{}
	PrintWarnings(analysis, buf, true)

	if buf.Len() != 0 {
		t.Errorf("PrintWarnings() with no warnings should not write anything, got: %v", buf.String())
	}
}

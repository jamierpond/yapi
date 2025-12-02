package validation

import (
	"strings"
	"testing"

	"yapi.run/cli/internal/utils"
)

func hasDiagnostic(diags []Diagnostic, substr string) bool {
	return utils.ContainsFunc(diags, func(d Diagnostic) bool {
		return strings.Contains(d.Message, substr)
	})
}

func TestAnalyzeConfig_ValidHTTP(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if len(a.Diagnostics) != 0 {
		t.Errorf("expected no diagnostics for valid HTTP config, got %d: %+v", len(a.Diagnostics), a.Diagnostics)
	}

	if a.Request == nil {
		t.Fatal("expected Request to be populated")
	}
}

func TestAnalyzeConfig_MissingURL(t *testing.T) {
	yaml := `yapi: v1
method: GET`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for missing URL")
	}

	if !hasDiagnostic(a.Diagnostics, "missing required field") {
		t.Errorf("expected 'missing required field' message, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_BadYAML(t *testing.T) {
	yaml := `yapi: v1
url: [invalid yaml`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for invalid YAML")
	}

	if !hasDiagnostic(a.Diagnostics, "invalid YAML") {
		t.Errorf("expected 'invalid YAML' message, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_BadGraphQL(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: |
  query { foo( }`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !hasDiagnostic(a.Diagnostics, "GraphQL syntax error") {
		t.Fatalf("expected GraphQL syntax error, got %+v", a.Diagnostics)
	}

	// Verify line number is set for GraphQL diagnostic
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "GraphQL") && d.Line < 0 {
			t.Errorf("expected GraphQL diagnostic to have line number set")
		}
	}
}

func TestAnalyzeConfig_ValidGraphQL(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: |
  query { foo }`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should have no GraphQL syntax errors
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "GraphQL") {
			t.Errorf("unexpected GraphQL diagnostic: %s", d.Message)
		}
	}
}

func TestAnalyzeConfig_BadJQ(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
jq_filter: .foo[`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !hasDiagnostic(a.Diagnostics, "JQ syntax error") {
		t.Fatalf("expected JQ syntax error, got %+v", a.Diagnostics)
	}

	// Verify line number is set for JQ diagnostic
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "JQ") && d.Line < 0 {
			t.Errorf("expected JQ diagnostic to have line number set")
		}
	}
}

func TestAnalyzeConfig_ValidJQ(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
jq_filter: .data.items[]`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should have no JQ syntax errors
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "JQ") {
			t.Errorf("unexpected JQ diagnostic: %s", d.Message)
		}
	}
}

func TestAnalyzeConfig_MissingVersion(t *testing.T) {
	yaml := `url: http://example.com
method: GET`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if len(a.Warnings) == 0 {
		t.Error("expected warning for missing yapi version")
	}

	found := false
	for _, w := range a.Warnings {
		if strings.Contains(w, "Missing") && strings.Contains(w, "v1") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected warning about missing version, got %+v", a.Warnings)
	}
}

func TestAnalyzeConfig_GRPCMissingRequirements(t *testing.T) {
	yaml := `yapi: v1
url: grpc://localhost:50051`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for gRPC missing service/rpc")
	}

	if !hasDiagnostic(a.Diagnostics, "service") {
		t.Errorf("expected error about missing service, got %+v", a.Diagnostics)
	}

	if !hasDiagnostic(a.Diagnostics, "rpc") {
		t.Errorf("expected error about missing rpc, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_TCPInvalidEncoding(t *testing.T) {
	yaml := `yapi: v1
url: tcp://localhost:9000
data: hello
encoding: invalid`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for invalid TCP encoding")
	}

	if !hasDiagnostic(a.Diagnostics, "unsupported TCP encoding") {
		t.Errorf("expected error about unsupported encoding, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_UnknownHTTPMethod(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: FOOBAR`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should have a warning, not an error
	if a.HasErrors() {
		t.Error("expected warning, not error, for unknown HTTP method")
	}

	var found bool
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, "unknown HTTP method") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected warning about unknown HTTP method")
	}
}

func TestAnalyzeConfig_GraphQLWithBody(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: query { foo }
body:
  key: value`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected error for graphql + body")
	}

	if !hasDiagnostic(a.Diagnostics, "cannot be used with") {
		t.Errorf("expected error about graphql/body conflict, got %+v", a.Diagnostics)
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name       string
		diags      []Diagnostic
		wantErrors bool
	}{
		{
			name:       "no diagnostics",
			diags:      nil,
			wantErrors: false,
		},
		{
			name: "only warnings",
			diags: []Diagnostic{
				{Severity: SeverityWarning, Message: "warning"},
			},
			wantErrors: false,
		},
		{
			name: "only info",
			diags: []Diagnostic{
				{Severity: SeverityInfo, Message: "info"},
			},
			wantErrors: false,
		},
		{
			name: "has errors",
			diags: []Diagnostic{
				{Severity: SeverityError, Message: "error"},
			},
			wantErrors: true,
		},
		{
			name: "mixed",
			diags: []Diagnostic{
				{Severity: SeverityWarning, Message: "warning"},
				{Severity: SeverityError, Message: "error"},
				{Severity: SeverityInfo, Message: "info"},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Analysis{Diagnostics: tt.diags}
			if got := a.HasErrors(); got != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", got, tt.wantErrors)
			}
		})
	}
}

func TestFindFieldLine(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
graphql: |
  query { foo }
jq_filter: .data`

	tests := []struct {
		field    string
		wantLine int
	}{
		{"yapi", 0},
		{"url", 1},
		{"method", 2},
		{"graphql", 3},
		{"jq_filter", 5},
		{"nonexistent", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := findFieldLine(yaml, tt.field)
			if got != tt.wantLine {
				t.Errorf("findFieldLine(%q) = %d, want %d", tt.field, got, tt.wantLine)
			}
		})
	}
}

func TestFindFieldLine_EmptyText(t *testing.T) {
	got := findFieldLine("", "field")
	if got != -1 {
		t.Errorf("findFieldLine with empty text = %d, want -1", got)
	}
}

func TestAnalyzeConfig_UnknownKeys(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
unknown_field: value
another_bad_key: 123`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should not have errors - unknown keys are warnings
	if a.HasErrors() {
		t.Error("expected no errors for unknown keys")
	}

	// Should have warning diagnostics about unknown keys
	var unknownKeyDiags []Diagnostic
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, "unknown key") {
			unknownKeyDiags = append(unknownKeyDiags, d)
		}
	}

	if len(unknownKeyDiags) < 2 {
		t.Errorf("expected at least 2 unknown key warnings, got %d", len(unknownKeyDiags))
	}

	// Check that line numbers are set correctly
	for _, d := range unknownKeyDiags {
		if d.Line < 0 {
			t.Errorf("expected line number for unknown key '%s', got %d", d.Field, d.Line)
		}
	}

	// Verify specific keys are detected
	if !hasDiagnostic(unknownKeyDiags, "unknown_field") {
		t.Errorf("expected warning about 'unknown_field', got %v", unknownKeyDiags)
	}

	if !hasDiagnostic(unknownKeyDiags, "another_bad_key") {
		t.Errorf("expected warning about 'another_bad_key', got %v", unknownKeyDiags)
	}
}

func TestAnalyzeConfig_NoUnknownKeys(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
headers:
  Authorization: Bearer token`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should have no warnings about unknown keys
	for _, w := range a.Warnings {
		if strings.Contains(w, "unknown key") {
			t.Errorf("unexpected unknown key warning: %s", w)
		}
	}
}

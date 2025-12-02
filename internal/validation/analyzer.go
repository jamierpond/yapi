package validation

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
)

// extractLineFromError attempts to extract a line number from YAML error messages.
// YAML errors often look like "line 22: cannot unmarshal..." - returns 0-indexed line or -1 if not found.
func extractLineFromError(errMsg string) int {
	re := regexp.MustCompile(`line (\d+):`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) >= 2 {
		if lineNum, err := strconv.Atoi(matches[1]); err == nil {
			return lineNum - 1 // Convert to 0-indexed
		}
	}
	return -1
}

// Diagnostic is the canonical diagnostic type that both CLI and LSP use.
type Diagnostic struct {
	Severity Severity
	Field    string // "url", "method", "graphql", "jq_filter", etc
	Message  string // human readable message

	// Optional position info. LSP uses it, CLI may ignore.
	Line int // 0-based, -1 if unknown
	Col  int // 0-based, -1 if unknown
}

// Analysis is the shared result type from analyzing a config.
type Analysis struct {
	Request     *domain.Request
	Diagnostics []Diagnostic
	Warnings    []string // parsed-level warnings like missing yapi: v1
}

// HasErrors returns true if there are any error-level diagnostics.
func (a *Analysis) HasErrors() bool {
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// AnalyzeConfigString is the single entrypoint for analyzing YAML config.
// Both CLI and LSP should call this function.
func AnalyzeConfigString(text string) (*Analysis, error) {
	parseRes, err := config.LoadFromString(text)
	if err != nil {
		// YAML parse error - no Request available
		// Try to extract line number from error message (e.g., "line 22: cannot unmarshal...")
		line := extractLineFromError(err.Error())
		diag := Diagnostic{
			Severity: SeverityError,
			Field:    "",
			Message:  fmt.Sprintf("invalid YAML: %v", err),
			Line:     line,
			Col:      0,
		}
		return &Analysis{Diagnostics: []Diagnostic{diag}}, nil
	}

	req := parseRes.Request
	var diags []Diagnostic

	// 1. Structural / semantic validation
	for _, iss := range ValidateRequest(req) {
		diags = append(diags, Diagnostic{
			Severity: iss.Severity,
			Field:    iss.Field,
			Message:  iss.Message,
			Line:     findFieldLine(text, iss.Field),
			Col:      0,
		})
	}

	// 2. GraphQL syntax validation
	diags = append(diags, ValidateGraphQLSyntax(text, req)...)

	// 3. JQ syntax validation
	diags = append(diags, ValidateJQSyntax(text, req)...)

	// 4. Unknown key detection
	diags = append(diags, validateUnknownKeys(text)...)

	return &Analysis{
		Request:     req,
		Diagnostics: diags,
		Warnings:    parseRes.Warnings,
	}, nil
}

// AnalyzeConfigFile loads a file and analyzes it.
func AnalyzeConfigFile(path string) (*Analysis, error) {
	parseRes, err := config.Load(path)
	if err != nil {
		diag := Diagnostic{
			Severity: SeverityError,
			Field:    "",
			Message:  fmt.Sprintf("failed to load config: %v", err),
			Line:     0,
			Col:      0,
		}
		return &Analysis{Diagnostics: []Diagnostic{diag}}, nil
	}

	// Re-read file to get text for line number detection
	// This is a bit redundant but keeps the API clean
	data, readErr := readFileForAnalysis(path)
	if readErr != nil {
		// Fall back to analysis without line numbers
		return analyzeRequest(parseRes.Request, "", parseRes.Warnings), nil
	}

	return analyzeRequest(parseRes.Request, string(data), parseRes.Warnings), nil
}

// analyzeRequest validates an already-parsed request.
func analyzeRequest(req *domain.Request, text string, warnings []string) *Analysis {
	var diags []Diagnostic

	// 1. Structural / semantic validation
	for _, iss := range ValidateRequest(req) {
		diags = append(diags, Diagnostic{
			Severity: iss.Severity,
			Field:    iss.Field,
			Message:  iss.Message,
			Line:     findFieldLine(text, iss.Field),
			Col:      0,
		})
	}

	// 2. GraphQL syntax validation
	diags = append(diags, ValidateGraphQLSyntax(text, req)...)

	// 3. JQ syntax validation
	diags = append(diags, ValidateJQSyntax(text, req)...)

	// 4. Unknown key detection
	diags = append(diags, validateUnknownKeys(text)...)

	return &Analysis{
		Request:     req,
		Diagnostics: diags,
		Warnings:    warnings,
	}
}

// readFileForAnalysis reads a file for analysis purposes.
func readFileForAnalysis(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// validateUnknownKeys checks for unknown keys in the YAML and returns warnings.
func validateUnknownKeys(text string) []Diagnostic {
	if text == "" {
		return nil
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(text), &raw); err != nil {
		return nil
	}

	unknownKeys := config.FindUnknownKeys(raw)
	var diags []Diagnostic
	for _, key := range unknownKeys {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Field:    key,
			Message:  fmt.Sprintf("unknown key '%s' will be ignored", key),
			Line:     findFieldLine(text, key),
			Col:      0,
		})
	}
	return diags
}

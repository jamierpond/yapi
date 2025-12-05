package validation

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	Warnings    []string           // parsed-level warnings like missing yapi: v1
	Chain       []config.ChainStep // Chain steps if this is a chain config
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

// JSONOutput is the JSON-serializable output for validation results.
type JSONOutput struct {
	Valid       bool             `json:"valid"`
	Diagnostics []JSONDiagnostic `json:"diagnostics"`
	Warnings    []string         `json:"warnings"`
}

// JSONDiagnostic is a JSON-serializable diagnostic.
type JSONDiagnostic struct {
	Severity string `json:"severity"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
}

// ToJSON converts the analysis to a JSON-serializable output.
func (a *Analysis) ToJSON() JSONOutput {
	diags := make([]JSONDiagnostic, 0, len(a.Diagnostics))
	for _, d := range a.Diagnostics {
		diags = append(diags, JSONDiagnostic{
			Severity: d.Severity.String(),
			Field:    d.Field,
			Message:  d.Message,
			Line:     d.Line,
			Col:      d.Col,
		})
	}

	warnings := a.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return JSONOutput{
		Valid:       !a.HasErrors(),
		Diagnostics: diags,
		Warnings:    warnings,
	}
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

	var diags []Diagnostic

	// Check if it is a chain
	if len(parseRes.Chain) > 0 {
		diags = append(diags, validateChain(text, parseRes.Chain)...)
		return &Analysis{
			Request:     nil,
			Chain:       parseRes.Chain,
			Diagnostics: diags,
			Warnings:    parseRes.Warnings,
		}, nil
	}

	req := parseRes.Request

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
	text := ""
	if readErr == nil {
		text = string(data)
	}

	// Check if it is a chain
	if len(parseRes.Chain) > 0 {
		var diags []Diagnostic
		diags = append(diags, validateChain(text, parseRes.Chain)...)
		return &Analysis{
			Request:     nil,
			Chain:       parseRes.Chain,
			Diagnostics: diags,
			Warnings:    parseRes.Warnings,
		}, nil
	}

	if readErr != nil {
		// Fall back to analysis without line numbers
		return analyzeRequest(parseRes.Request, "", parseRes.Warnings), nil
	}

	return analyzeRequest(parseRes.Request, text, parseRes.Warnings), nil
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

// varRefRegex captures variable references ($var and ${var}) for chain validation
var varRefRegex = regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z0-9_\-\.]+)`)

// validateChain validates chain configuration
func validateChain(text string, chain []config.ChainStep) []Diagnostic {
	var diags []Diagnostic
	definedSteps := make(map[string]bool)

	for i, step := range chain {
		// 1. Check name is present
		if step.Name == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Message:  fmt.Sprintf("step #%d missing 'name'", i+1),
				Line:     -1,
				Col:      0,
			})
		} else if definedSteps[step.Name] {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("duplicate step name '%s'", step.Name),
				Line:     -1,
				Col:      0,
			})
		}

		// 2. Check URL is present
		if step.URL == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("step '%s' missing 'url'", step.Name),
				Line:     -1,
				Col:      0,
			})
		}

		// 3. Check for references to future steps
		diags = append(diags, scanForUndefinedRefs(step.URL, definedSteps, step.Name, "url")...)

		// Check Headers
		for _, v := range step.Headers {
			diags = append(diags, scanForUndefinedRefs(v, definedSteps, step.Name, "headers")...)
		}

		// Check Body values (if they are strings)
		for k, v := range step.Body {
			if s, ok := v.(string); ok {
				diags = append(diags, scanForUndefinedRefs(s, definedSteps, step.Name, fmt.Sprintf("body.%s", k))...)
			}
		}

		// Check JSON field
		if step.JSON != "" {
			diags = append(diags, scanForUndefinedRefs(step.JSON, definedSteps, step.Name, "json")...)
		}

		// Check Variables
		for k, v := range step.Variables {
			if s, ok := v.(string); ok {
				diags = append(diags, scanForUndefinedRefs(s, definedSteps, step.Name, fmt.Sprintf("variables.%s", k))...)
			}
		}

		// 4. Add to defined scope
		if step.Name != "" {
			definedSteps[step.Name] = true
		}
	}
	return diags
}

// scanForUndefinedRefs checks a value string for references to undefined steps
func scanForUndefinedRefs(value string, definedSteps map[string]bool, currentStep, fieldName string) []Diagnostic {
	var diags []Diagnostic
	matches := varRefRegex.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		var key string
		if strings.HasPrefix(match[0], "${") {
			key = match[1]
		} else {
			key = match[2]
		}

		// Only check chain references (containing dot)
		if strings.Contains(key, ".") {
			parts := strings.Split(key, ".")
			refStep := parts[0]

			if !definedSteps[refStep] {
				msg := fmt.Sprintf("step '%s' references '%s' before it is defined", currentStep, refStep)
				if refStep == currentStep {
					msg = fmt.Sprintf("step '%s' cannot reference itself", currentStep)
				}

				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Field:    fmt.Sprintf("%s.%s", currentStep, fieldName),
					Message:  msg,
					Line:     -1,
					Col:      0,
				})
			}
		}
	}
	return diags
}

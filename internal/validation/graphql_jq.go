package validation

import (
	"strings"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/itchyny/gojq"
	"yapi.run/cli/internal/domain"
)

// ValidateGraphQLSyntax validates the GraphQL query syntax if present.
func ValidateGraphQLSyntax(fullYaml string, req *domain.Request) []Diagnostic {
	q, ok := req.Metadata["graphql_query"]
	if !ok || q == "" {
		return nil
	}

	src := source.NewSource(&source.Source{
		Body: []byte(q),
		Name: "GraphQL Query",
	})

	_, err := parser.Parse(parser.ParseParams{Source: src})
	if err == nil {
		return nil
	}

	line := findFieldLine(fullYaml, "graphql")
	// GraphQL content typically starts on line after "graphql: |"
	if line >= 0 {
		line++
	}

	return []Diagnostic{{
		Severity: SeverityError,
		Field:    "graphql",
		Message:  "GraphQL syntax error: " + err.Error(),
		Line:     line,
		Col:      0,
	}}
}

// ValidateJQSyntax validates the jq filter syntax if present.
func ValidateJQSyntax(fullYaml string, req *domain.Request) []Diagnostic {
	f, ok := req.Metadata["jq_filter"]
	if !ok || strings.TrimSpace(f) == "" {
		return nil
	}

	_, err := gojq.Parse(f)
	if err == nil {
		return nil
	}

	line := findFieldLine(fullYaml, "jq_filter")

	return []Diagnostic{{
		Severity: SeverityError,
		Field:    "jq_filter",
		Message:  "JQ syntax error: " + err.Error(),
		Line:     line,
		Col:      0,
	}}
}

// findFieldLine finds the line number (0-based) of a YAML field.
// Returns -1 if not found or if text is empty.
func findFieldLine(text, field string) int {
	if field == "" || text == "" {
		return -1
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), field+":") {
			return i
		}
	}
	return -1
}

// ValidateChainAssertions validates JQ syntax for all assertions in chain steps.
func ValidateChainAssertions(text string, assertions []string, stepName string) []Diagnostic {
	var diags []Diagnostic

	for _, assertion := range assertions {
		_, err := gojq.Parse(assertion)
		if err != nil {
			// Find the line where this assertion appears
			line := findValueInTextForAssertion(text, assertion)

			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    stepName + ".assert",
				Message:  "JQ syntax error: " + err.Error(),
				Line:     line,
				Col:      0,
			})
		}
	}

	return diags
}

// findValueInTextForAssertion finds the line where an assertion string appears
func findValueInTextForAssertion(text, assertion string) int {
	if text == "" || assertion == "" {
		return -1
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Check if line contains the assertion (with possible quotes or dashes)
		if strings.Contains(line, assertion) {
			return i
		}
	}
	return -1
}

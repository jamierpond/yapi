package validation

import (
	"regexp"
	"strings"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
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
// Uses YAML parsing for accurate position, with regex fallback.
// Returns -1 if not found or if text is empty.
func findFieldLine(text, field string) int {
	if field == "" || text == "" {
		return -1
	}

	// Try YAML node parsing first for accuracy
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(text), &node); err == nil {
		if line := findFieldInNode(&node, field); line >= 0 {
			return line
		}
	}

	// Fallback: use regex to match field as a complete word followed by colon
	// This handles cases where YAML parsing succeeds but field is nested differently
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(field) + `\s*:`)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if pattern.MatchString(line) {
			return i
		}
	}
	return -1
}

// findFieldInNode recursively searches a YAML node tree for a field name.
func findFieldInNode(node *yaml.Node, field string) int {
	if node == nil {
		return -1
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if line := findFieldInNode(child, field); line >= 0 {
				return line
			}
		}
	case yaml.MappingNode:
		// Content alternates between keys and values
		for i := 0; i < len(node.Content)-1; i += 2 {
			keyNode := node.Content[i]
			if keyNode.Value == field {
				return keyNode.Line - 1 // yaml.Node lines are 1-based
			}
			// Also search in the value node (for nested fields)
			if line := findFieldInNode(node.Content[i+1], field); line >= 0 {
				return line
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if line := findFieldInNode(child, field); line >= 0 {
				return line
			}
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

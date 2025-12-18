package validation

import (
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/itchyny/gojq"
	"yapi.run/cli/internal/domain"
)

// ValidateGraphQLSyntax validates the GraphQL query syntax if present.
func ValidateGraphQLSyntax(file *ast.File, req *domain.Request) []Diagnostic {
	q, ok := req.Metadata["graphql_query"]
	if !ok || q == "" {
		return nil
	}

	// 1. Validate Syntax
	src := source.NewSource(&source.Source{
		Body: []byte(q),
		Name: "GraphQL Query",
	})
	_, err := parser.Parse(parser.ParseParams{Source: src})
	if err == nil {
		return nil
	}

	// 2. Find Location in AST
	line, col := findKeyLocationInDocs(file, "graphql")
	// If the graphql query is a block scalar ( | ), the error is likely on the next line
	// However, we just point to the start of the field for now or the scalar value
	if valNode := findValueNode(file, "graphql"); valNode != nil {
		line = valNode.GetToken().Position.Line - 1
		col = valNode.GetToken().Position.Column - 1
	}

	return []Diagnostic{{
		Severity: SeverityError,
		Field:    "graphql",
		Message:  "GraphQL syntax error: " + err.Error(),
		Line:     line,
		Col:      col,
	}}
}

// ValidateJQSyntax validates the jq filter syntax if present.
func ValidateJQSyntax(file *ast.File, req *domain.Request) []Diagnostic {
	f, ok := req.Metadata["jq_filter"]
	if !ok || strings.TrimSpace(f) == "" {
		return nil
	}

	_, err := gojq.Parse(f)
	if err == nil {
		return nil
	}

	line, col := findKeyLocationInDocs(file, "jq_filter")
	if valNode := findValueNode(file, "jq_filter"); valNode != nil {
		line = valNode.GetToken().Position.Line - 1
		col = valNode.GetToken().Position.Column - 1
	}

	return []Diagnostic{{
		Severity: SeverityError,
		Field:    "jq_filter",
		Message:  "JQ syntax error: " + err.Error(),
		Line:     line,
		Col:      col,
	}}
}

// validateAssertionsWithAST finds assertions in top-level config
func validateAssertionsWithAST(file *ast.File, assertions []string, prefix string, section string) []Diagnostic {
	// Find 'expect' -> 'assert' -> section node
	expectNode := findValueNode(file, "expect")
	if expectNode == nil {
		return validateAssertionsFallback(assertions, prefix, 0)
	}

	assertNode := findKeyInMap(expectNode, "assert")
	if assertNode == nil {
		return validateAssertionsFallback(assertions, prefix, 0)
	}

	targetNode := findKeyInMap(assertNode, section)

	return validateAssertionsInNode(targetNode, assertions, prefix)
}

// validateAssertionsInStep validates assertions inside a specific chain step node
func validateAssertionsInStep(stepNode ast.Node, assertions []string, stepName string, section string) []Diagnostic {
	if stepNode == nil {
		return nil
	}

	expectNode := findKeyInMap(stepNode, "expect")
	assertNode := findKeyInMap(expectNode, "assert")
	targetNode := findKeyInMap(assertNode, section)

	return validateAssertionsInNode(targetNode, assertions, stepName+".assert")
}

func validateAssertionsInNode(node ast.Node, assertions []string, errorFieldPrefix string) []Diagnostic {
	if node == nil {
		return nil
	}

	var diags []Diagnostic

	// The node could be a Sequence (list of assertions) or just one?
	// Usually expect.assert.body is a sequence.
	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		return nil
	}

	for _, val := range seq.Values {
		strNode, ok := val.(*ast.StringNode)
		if !ok {
			continue
		}

		// Validate this specific assertion string
		_, err := gojq.Parse(strNode.Value)
		if err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    errorFieldPrefix,
				Message:  "JQ syntax error: " + err.Error(),
				Line:     strNode.GetToken().Position.Line - 1,
				Col:      strNode.GetToken().Position.Column - 1,
			})
		}
	}
	return diags
}

// Fallback if AST lookup fails completely
func validateAssertionsFallback(assertions []string, prefix string, defaultLine int) []Diagnostic {
	var diags []Diagnostic
	for _, a := range assertions {
		if _, err := gojq.Parse(a); err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    prefix,
				Message:  err.Error(),
				Line:     defaultLine,
			})
		}
	}
	return diags
}

// AST Helpers

func findKeyLocationInDocs(file *ast.File, key string) (int, int) {
	if len(file.Docs) == 0 {
		return 0, 0
	}
	if mapping, ok := file.Docs[0].Body.(*ast.MappingNode); ok {
		for _, kv := range mapping.Values {
			if k, ok := kv.Key.(*ast.StringNode); ok && k.Value == key {
				return k.GetToken().Position.Line - 1, k.GetToken().Position.Column - 1
			}
		}
	}
	return 0, 0
}

func findValueNode(file *ast.File, key string) ast.Node {
	if len(file.Docs) == 0 {
		return nil
	}
	return findKeyInMap(file.Docs[0].Body, key)
}

func findKeyInMap(node ast.Node, key string) ast.Node {
	if node == nil {
		return nil
	}
	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		return nil
	}
	for _, kv := range mapping.Values {
		if k, ok := kv.Key.(*ast.StringNode); ok && k.Value == key {
			return kv.Value
		}
	}
	return nil
}

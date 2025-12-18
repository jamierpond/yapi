// Package validation provides config analysis and diagnostics.
package validation

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/vars"
)

// Diagnostic is the canonical diagnostic type.
type Diagnostic struct {
	Severity Severity
	Field    string
	Message  string
	Line     int // 0-based
	Col      int // 0-based
}

// Analysis is the shared result type.
type Analysis struct {
	Request     *domain.Request
	Diagnostics []Diagnostic
	Warnings    []string
	Chain       []config.ChainStep
	Base        *config.ConfigV1
	Expect      config.Expectation
	// AST is preserved for advanced tooling/LSP usages if needed later
	AST *ast.File
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

// AnalyzeConfigString entrypoint.
func AnalyzeConfigString(text string) (*Analysis, error) {
	return analyzeCommon([]byte(text))
}

// AnalyzeConfigFile entrypoint.
func AnalyzeConfigFile(path string) (*Analysis, error) {
	var data []byte
	var err error

	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path) //nolint:gosec // user-provided config path
	}

	if err != nil {
		return &Analysis{Diagnostics: []Diagnostic{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("failed to read config: %v", err),
		}}}, nil
	}

	return analyzeCommon(data)
}

func analyzeCommon(data []byte) (*Analysis, error) {
	// 1. Parse into Domain/Structs (Logic) using existing loader
	// We keep using the config loader for logic because it handles defaults, merging, etc.
	parseRes, err := config.LoadFromString(string(data))
	if err != nil {
		// Attempt to get line number from yaml error
		line := extractLineFromError(err.Error())
		return &Analysis{Diagnostics: []Diagnostic{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("invalid YAML: %v", err),
			Line:     line,
		}}}, nil
	}

	// 2. Parse into AST (Location/Syntax)
	// We use Mode 0 (default)
	fileNode, err := parser.ParseBytes(data, 0)
	if err != nil {
		// If AST parsing fails but structural parsing succeeded (unlikely), fallback
		return &Analysis{Diagnostics: []Diagnostic{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("syntax error: %v", err),
		}}}, nil
	}

	return analyzeParsed(fileNode, parseRes, string(data)), nil
}

// analyzeParsed performs validation using both the struct data (logic) and AST (location).
func analyzeParsed(fileNode *ast.File, parseRes *config.ParseResult, text string) *Analysis {
	var diags []Diagnostic

	// Helper to find location of a top-level field
	getLocation := func(field string) (int, int) {
		return findFieldLocation(fileNode, field)
	}

	// Chain config validation
	if len(parseRes.Chain) > 0 {
		diags = append(diags, validateChain(fileNode, parseRes.Base, parseRes.Chain)...)
		diags = append(diags, validateEnvVars(fileNode, text)...)
		return &Analysis{
			Chain:       parseRes.Chain,
			Base:        parseRes.Base,
			Diagnostics: diags,
			Warnings:    parseRes.Warnings,
			AST:         fileNode,
		}
	}

	// Single Request Validation
	req := parseRes.Request
	for _, iss := range ValidateRequest(req) {
		line, col := getLocation(iss.Field)
		diags = append(diags, Diagnostic{
			Severity: iss.Severity,
			Field:    iss.Field,
			Message:  iss.Message,
			Line:     line,
			Col:      col,
		})
	}

	// Syntax Validators
	diags = append(diags, ValidateGraphQLSyntax(fileNode, req)...)
	diags = append(diags, ValidateJQSyntax(fileNode, req)...)
	diags = append(diags, validateUnknownKeys(fileNode)...)
	diags = append(diags, validateEnvVars(fileNode, text)...) // AST-based env var finder

	// Assertion Validation
	if len(parseRes.Expect.Assert.Body) > 0 {
		// Locate the 'expect' -> 'assert' -> 'body' node or simple list
		diags = append(diags, validateAssertionsWithAST(fileNode, parseRes.Expect.Assert.Body, "", "body")...)
	}
	if len(parseRes.Expect.Assert.Headers) > 0 {
		diags = append(diags, validateAssertionsWithAST(fileNode, parseRes.Expect.Assert.Headers, "", "headers")...)
	}

	return &Analysis{
		Request:     req,
		Diagnostics: diags,
		Warnings:    parseRes.Warnings,
		Expect:      parseRes.Expect,
		Base:        parseRes.Base,
		AST:         fileNode,
	}
}

// extractLineFromError attempts to parse "line X:" from yaml errors.
func extractLineFromError(errMsg string) int {
	// Simple heuristic, regex matching "line \d+"
	// In a real scenario, use regex. Keeping it simple here as per original.
	if idx := strings.Index(errMsg, "line "); idx != -1 {
		rest := errMsg[idx+5:]
		if end := strings.IndexAny(rest, ": \t"); end != -1 {
			if val, err := strconv.Atoi(rest[:end]); err == nil {
				return val - 1
			}
		}
	}
	return 0
}

// AST Traversal Helpers

// findFieldLocation finds the (line, col) of a key in the top-level mapping.
func findFieldLocation(file *ast.File, keyName string) (int, int) {
	if len(file.Docs) == 0 {
		return 0, 0
	}
	body := file.Docs[0].Body
	mapping, ok := body.(*ast.MappingNode)
	if !ok {
		return 0, 0
	}

	for _, kv := range mapping.Values {
		if key, ok := kv.Key.(*ast.StringNode); ok && key.Value == keyName {
			// Convert to 0-based
			return key.GetToken().Position.Line - 1, key.GetToken().Position.Column - 1
		}
	}
	return 0, 0
}

// validateUnknownKeys uses AST to find keys not in the allowed list.
func validateUnknownKeys(file *ast.File) []Diagnostic {
	if len(file.Docs) == 0 {
		return nil
	}
	body := file.Docs[0].Body
	mapping, ok := body.(*ast.MappingNode)
	if !ok {
		return nil
	}

	var diags []Diagnostic
	for _, kv := range mapping.Values {
		keyNode, ok := kv.Key.(*ast.StringNode)
		if !ok {
			continue
		}

		key := keyNode.Value
		if !config.IsKnownKey(key) {
			tk := keyNode.GetToken()
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    key,
				Message:  fmt.Sprintf("unknown key '%s' will be ignored", key),
				Line:     tk.Position.Line - 1,
				Col:      tk.Position.Column - 1,
			})
		}
	}
	return diags
}

// validateChain uses AST to validate chain steps.
func validateChain(file *ast.File, base *config.ConfigV1, chain []config.ChainStep) []Diagnostic {
	var diags []Diagnostic
	definedSteps := make(map[string]bool)

	// 1. Locate the "chain" node in AST
	var chainSeq *ast.SequenceNode

	if len(file.Docs) > 0 {
		if mapping, ok := file.Docs[0].Body.(*ast.MappingNode); ok {
			for _, kv := range mapping.Values {
				if k, ok := kv.Key.(*ast.StringNode); ok && k.Value == "chain" {
					if seq, ok := kv.Value.(*ast.SequenceNode); ok {
						chainSeq = seq
					}
					break
				}
			}
		}
	}

	for i, step := range chain {
		// Attempt to find the AST node for this step
		var stepNode *ast.MappingNode
		var stepLine, stepCol int

		if chainSeq != nil && i < len(chainSeq.Values) {
			if m, ok := chainSeq.Values[i].(*ast.MappingNode); ok {
				stepNode = m
				stepLine = m.GetToken().Position.Line - 1
				stepCol = m.GetToken().Position.Column - 1
			}
		}

		// Validation 1: Missing Name
		if step.Name == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Message:  fmt.Sprintf("step #%d missing 'name'", i+1),
				Line:     stepLine,
				Col:      stepCol,
			})
		} else if definedSteps[step.Name] {
			// Find specific location of "name" key if possible
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("duplicate step name '%s'", step.Name),
				Line:     stepLine,
				Col:      stepCol,
			})
		}

		// Validation 2: Missing URL
		hasURL := step.URL != "" || (base != nil && base.URL != "")
		if !hasURL {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("step '%s' missing 'url'", step.Name),
				Line:     stepLine,
				Col:      stepCol,
			})
		}

		// Validation 3: Variable Refs (Recursively walk the stepNode)
		if stepNode != nil {
			diags = append(diags, validateUndefinedRefsInNode(stepNode, definedSteps, step.Name)...)
		}

		// Validation 4: JQ Assertions
		// We pass the AST node to find exact lines of assertion strings
		if step.Expect.Assert.Body != nil {
			diags = append(diags, validateAssertionsInStep(stepNode, step.Expect.Assert.Body, step.Name, "body")...)
		}
		if step.Expect.Assert.Headers != nil {
			diags = append(diags, validateAssertionsInStep(stepNode, step.Expect.Assert.Headers, step.Name, "headers")...)
		}

		if step.Name != "" {
			definedSteps[step.Name] = true
		}
	}

	return diags
}

// validateUndefinedRefsInNode recursively walks an AST node looking for string values with ${...}
func validateUndefinedRefsInNode(node ast.Node, definedSteps map[string]bool, currentStep string) []Diagnostic {
	var diags []Diagnostic
	walkNode(node, func(n ast.Node) {
		// We only care about Scalar String nodes
		strNode, ok := n.(*ast.StringNode)
		if !ok {
			return
		}

		// Check if it matches variable pattern
		val := strNode.Value
		matches := vars.ChainVar.FindAllString(val, -1)

		for _, match := range matches {
			// match is like "${step.field}" or "$step.field"
			clean := strings.TrimPrefix(match, "$")
			clean = strings.TrimPrefix(clean, "{")
			clean = strings.TrimSuffix(clean, "}")

			parts := strings.Split(clean, ".")
			if len(parts) < 2 {
				continue
			}
			refStep := parts[0]

			if !definedSteps[refStep] {
				msg := fmt.Sprintf("step '%s' references '%s' before it is defined", currentStep, refStep)
				if refStep == currentStep {
					msg = fmt.Sprintf("step '%s' cannot reference itself", currentStep)
				}

				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Message:  msg,
					Line:     strNode.GetToken().Position.Line - 1,
					Col:      strNode.GetToken().Position.Column - 1,
				})
			}
		}
	})

	return diags
}

// validateEnvVars uses AST to find environment variables.
// This is much more accurate than regex on raw text as it respects comments and block scalars.
func validateEnvVars(file *ast.File, rawText string) []Diagnostic {
	var diags []Diagnostic

	// Define a visitor that looks at string scalars
	for _, doc := range file.Docs {
		walkNode(doc.Body, func(n ast.Node) {
			strNode, ok := n.(*ast.StringNode)
			if !ok {
				return
			}

			// Check for env vars: $VAR or ${VAR}
			// Note: We use the *Raw* value to check if it looks like a variable in the source
			// ignoring what the YAML parser interpreted it as (though usually they align).
			val := strNode.Value

			matches := vars.EnvOnly.FindAllStringSubmatch(val, -1)
			for _, match := range matches {
				// Ensure it's not a chain var (containing dot)
				fullMatch := match[0]
				if strings.Contains(fullMatch, ".") {
					continue
				}

				// Extract var name
				var varName string
				if match[1] != "" {
					varName = match[1] // ${VAR}
				} else {
					varName = match[2] // $VAR
				}

				// Check definitions
				if os.Getenv(varName) == "" {
					tk := strNode.GetToken()
					diags = append(diags, Diagnostic{
						Severity: SeverityWarning,
						Field:    varName,
						Message:  fmt.Sprintf("environment variable '%s' is not defined", varName),
						Line:     tk.Position.Line - 1,
						Col:      tk.Position.Column - 1,
					})
				}
			}
		})
	}

	return diags
}

// walkNode recursively walks an AST node and applies a function to each node.
func walkNode(node ast.Node, fn func(ast.Node)) {
	if node == nil {
		return
	}

	fn(node)

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, kv := range n.Values {
			walkNode(kv.Key, fn)
			walkNode(kv.Value, fn)
		}
	case *ast.MappingValueNode:
		walkNode(n.Key, fn)
		walkNode(n.Value, fn)
	case *ast.SequenceNode:
		for _, v := range n.Values {
			walkNode(v, fn)
		}
	case *ast.AnchorNode:
		walkNode(n.Value, fn)
	case *ast.AliasNode:
		// No children to walk
	case *ast.StringNode:
		// Leaf node
	case *ast.IntegerNode:
		// Leaf node
	case *ast.FloatNode:
		// Leaf node
	case *ast.BoolNode:
		// Leaf node
	case *ast.NullNode:
		// Leaf node
	case *ast.TagNode:
		walkNode(n.Value, fn)
	}
}

// findFieldLine finds the line number (0-based) of a YAML field using AST parsing.
// Returns -1 if not found or if text is empty.
// This is kept for backwards compatibility with tests.
func findFieldLine(text, field string) int {
	if field == "" || text == "" {
		return -1
	}

	fileNode, err := parser.ParseBytes([]byte(text), 0)
	if err != nil {
		return -1
	}

	if len(fileNode.Docs) == 0 {
		return -1
	}
	body := fileNode.Docs[0].Body
	mapping, ok := body.(*ast.MappingNode)
	if !ok {
		return -1
	}

	for _, kv := range mapping.Values {
		if key, ok := kv.Key.(*ast.StringNode); ok && key.Value == field {
			return key.GetToken().Position.Line - 1
		}
	}
	return -1
}

// EnvVarInfo holds information about an env var reference for hover/diagnostics
type EnvVarInfo struct {
	Name       string
	Value      string // Empty if not defined
	IsDefined  bool
	Line       int
	Col        int
	StartIndex int
	EndIndex   int
}

// FindEnvVarRefs finds all environment variable references in text
// This is kept for LSP hover functionality and uses regex on raw text
// to avoid needing to parse the YAML AST in the langserver.
func FindEnvVarRefs(text string) []EnvVarInfo {
	var refs []EnvVarInfo
	lines := strings.Split(text, "\n")

	// Track if we're inside a graphql block (which uses $var syntax for GraphQL variables)
	inGraphQLBlock := false
	graphqlIndent := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for graphql: field start
		if strings.HasPrefix(trimmed, "graphql:") {
			inGraphQLBlock = true
			// Find the indentation of the graphql key
			graphqlIndent = len(line) - len(strings.TrimLeft(line, " \t"))
			continue
		}

		// If we're in a graphql block, check if we've exited it
		if inGraphQLBlock {
			// Empty lines stay in block
			if trimmed == "" {
				continue
			}
			// Calculate current line's indentation
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			// If current indentation is <= graphql key's indentation and line has content,
			// we've exited the block (unless it's a continuation like |)
			if currentIndent <= graphqlIndent && !strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, ">") {
				inGraphQLBlock = false
			} else {
				// Still in graphql block - skip $var matching (GraphQL variables)
				continue
			}
		}

		matches := vars.EnvOnly.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			// match[0:2] = full match, match[2:4] = ${VAR} capture, match[4:6] = $VAR capture
			fullStart, fullEnd := match[0], match[1]
			fullMatch := line[fullStart:fullEnd]

			// Skip if this looks like a chain reference (contains a dot after the var name)
			// Check the character after the match
			if fullEnd < len(line) && line[fullEnd] == '.' {
				continue
			}

			var varName string
			if match[2] != -1 {
				// ${VAR} style
				varName = line[match[2]:match[3]]
			} else if match[4] != -1 {
				// $VAR style
				varName = line[match[4]:match[5]]
			}

			if varName == "" {
				continue
			}

			// Check if it's actually an env var (not a chain ref)
			// Chain refs have dots like ${step.field}
			if strings.Contains(fullMatch, ".") {
				continue
			}

			// Skip JQ variables (start with underscore, e.g., $_headers, $_body)
			if strings.HasPrefix(varName, "_") {
				continue
			}

			value := os.Getenv(varName)
			refs = append(refs, EnvVarInfo{
				Name:       varName,
				Value:      value,
				IsDefined:  value != "",
				Line:       lineNum,
				Col:        fullStart,
				StartIndex: fullStart,
				EndIndex:   fullEnd,
			})
		}
	}
	return refs
}

// RedactValue redacts a value for display, showing only first/last chars
func RedactValue(value string) string {
	if value == "" {
		return "(empty)"
	}
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

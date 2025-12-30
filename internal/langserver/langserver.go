// Package langserver implements an LSP server for yapi config files.
package langserver

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/compiler"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/validation"
	"yapi.run/cli/internal/vars"
)

const lsName = "yapi language server"

var (
	version = "0.0.1"
	handler protocol.Handler
	docs    = make(map[protocol.DocumentUri]*document)
)

type document struct {
	URI         protocol.DocumentUri
	Text        string
	ProjectRoot string                  // Path to project root (if found)
	Project     *config.ProjectConfigV1 // Project config (if found)
}

// Run starts the yapi language server over stdio.
func Run() {
	commonlog.Configure(1, nil)

	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		SetTrace:               setTrace,
		TextDocumentDidOpen:    textDocumentDidOpen,
		TextDocumentDidChange:  textDocumentDidChange,
		TextDocumentDidClose:   textDocumentDidClose,
		TextDocumentDidSave:    textDocumentDidSave,
		TextDocumentCompletion: textDocumentCompletion,
		TextDocumentHover:      textDocumentHover,
		TextDocumentDefinition: textDocumentDefinition,
	}

	srv := server.NewServer(&handler, lsName, false)
	_ = srv.RunStdio()
}

func initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	syncKind := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{
		OpenClose: boolPtr(true),
		Change:    &syncKind,
		Save: &protocol.SaveOptions{
			IncludeText: boolPtr(true),
		},
	}

	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{":", " ", "\n"},
	}

	capabilities.HoverProvider = true

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(ctx *glsp.Context) error {
	return nil
}

func setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	return nil
}

func textDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	text := params.TextDocument.Text

	doc := &document{
		URI:  uri,
		Text: text,
	}

	// Try to find and load project config
	if filePath := uriToPath(uri); filePath != "" {
		dirPath := filepath.Dir(filePath)
		if projectRoot, err := config.FindProjectRoot(dirPath); err == nil {
			doc.ProjectRoot = projectRoot
			if project, err := config.LoadProject(projectRoot); err == nil {
				doc.Project = project
			}
		}
	}

	docs[uri] = doc
	validateAndNotify(ctx, uri, text)
	return nil
}

func textDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI

	// With TextDocumentSyncKindFull, we get the full text in each change
	if len(params.ContentChanges) > 0 {
		text := params.ContentChanges[len(params.ContentChanges)-1].(protocol.TextDocumentContentChangeEventWhole).Text

		if doc, ok := docs[uri]; ok {
			// Update text but preserve project context
			doc.Text = text
		} else {
			// Create new document if it doesn't exist
			doc := &document{
				URI:  uri,
				Text: text,
			}
			// Try to load project context
			if filePath := uriToPath(uri); filePath != "" {
				dirPath := filepath.Dir(filePath)
				if projectRoot, err := config.FindProjectRoot(dirPath); err == nil {
					doc.ProjectRoot = projectRoot
					if project, err := config.LoadProject(projectRoot); err == nil {
						doc.Project = project
					}
				}
			}
			docs[uri] = doc
		}

		validateAndNotify(ctx, uri, text)
	}

	return nil
}

func textDocumentDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	delete(docs, params.TextDocument.URI)
	return nil
}

func textDocumentDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	if params.Text != nil {
		uri := params.TextDocument.URI
		text := *params.Text

		if doc, ok := docs[uri]; ok {
			doc.Text = text
		}

		validateAndNotify(ctx, uri, text)
	}
	return nil
}

func validateAndNotify(ctx *glsp.Context, uri protocol.DocumentUri, text string) {
	// Check if this is a project config file
	if filePath := uriToPath(uri); filePath != "" {
		fileName := filepath.Base(filePath)
		if fileName == "yapi.config.yml" || fileName == "yapi.config.yaml" {
			// This is a project config file - validate it
			validateProjectConfig(ctx, uri, text, filePath)
			return
		}
	}

	// Get document to access project context
	doc, ok := docs[uri]
	var analysis *validation.Analysis
	var err error

	// Use project-aware validation if available
	if ok && doc.Project != nil {
		analysis, err = validation.AnalyzeConfigStringWithProject(text, doc.Project, doc.ProjectRoot)
	} else {
		analysis, err = validation.AnalyzeConfigString(text)
	}

	if err != nil || analysis == nil {
		analysis, err = validation.AnalyzeConfigString(text)
	}
	if err != nil {
		// Catastrophic error - send one diagnostic and bail
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []protocol.Diagnostic{{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 1},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr("yapi"),
				Message:  "internal validation error: " + err.Error(),
			}},
		})
		return
	}

	// Initialize to empty slice, not nil, so JSON serializes as [] not null
	diagnostics := []protocol.Diagnostic{}

	// Config-level warnings (missing yapi: v1 etc)
	for _, w := range analysis.Warnings {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 100},
			},
			Severity: ptr(protocol.DiagnosticSeverityWarning),
			Source:   ptr("yapi"),
			Message:  w,
		})
	}

	// Analyzer diagnostics
	for _, d := range analysis.Diagnostics {
		line := protocol.UInteger(0)
		char := protocol.UInteger(0)
		if d.Line >= 0 {
			line = protocol.UInteger(d.Line)
		}
		if d.Col >= 0 {
			char = protocol.UInteger(d.Col)
		}

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: char},
				End:   protocol.Position{Line: line, Character: 100},
			},
			Severity: ptr(severityToProtocol(d.Severity)),
			Source:   ptr("yapi"),
			Message:  d.Message,
		})
	}

	// Compiler parity check - run the compiler with mock resolver for additional validation
	// Skip for chain configs (they require different handling)
	if analysis.Request != nil && len(analysis.Chain) == 0 && !analysis.HasErrors() {
		var resolver vars.Resolver = vars.MockResolver
		compiled := compiler.Compile(analysis.Base, resolver)
		for _, err := range compiled.Errors {
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 100},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr("yapi-compiler"),
				Message:  err.Error(),
			})
		}
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func validateProjectConfig(ctx *glsp.Context, uri protocol.DocumentUri, text string, filePath string) {
	diagnostics := []protocol.Diagnostic{}

	// Parse YAML first to catch syntax errors
	var rawConfig map[string]any
	if err := yaml.Unmarshal([]byte(text), &rawConfig); err != nil {
		// Try to extract line number from error message
		// Error format: "yaml: line X: ..."
		line := protocol.UInteger(0)
		errMsg := err.Error()
		if strings.Contains(errMsg, "line ") {
			var lineNum int
			if _, scanErr := fmt.Sscanf(errMsg, "yaml: line %d:", &lineNum); scanErr == nil && lineNum > 0 {
				line = protocol.UInteger(lineNum - 1) // LSP uses 0-indexed lines
			}
		}

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: 0},
				End:   protocol.Position{Line: line, Character: 100},
			},
			Severity: ptr(protocol.DiagnosticSeverityError),
			Source:   ptr("yapi"),
			Message:  fmt.Sprintf("YAML syntax error: %v", err),
		})
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
		return
	}

	// Check required fields
	if _, ok := rawConfig["yapi"]; !ok {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 100},
			},
			Severity: ptr(protocol.DiagnosticSeverityError),
			Source:   ptr("yapi"),
			Message:  "Missing required field 'yapi' (e.g., yapi: v1)",
		})
	}

	// Check if default_environment references a valid environment
	if defaultEnv, ok := rawConfig["default_environment"].(string); ok && defaultEnv != "" {
		if envs, hasEnvs := rawConfig["environments"].(map[string]any); hasEnvs {
			if _, exists := envs[defaultEnv]; !exists {
				// Get line number for default_environment
				line := findFieldLineInText(text, "default_environment")
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{Line: protocol.UInteger(line), Character: 0},
						End:   protocol.Position{Line: protocol.UInteger(line), Character: 100},
					},
					Severity: ptr(protocol.DiagnosticSeverityError),
					Source:   ptr("yapi"),
					Message:  fmt.Sprintf("default_environment '%s' not found in environments", defaultEnv),
				})
			}
		}
	}

	// Try to load the full project config for additional validation
	projectRoot := filepath.Dir(filePath)
	_, err := config.LoadProject(projectRoot)

	if err != nil && len(diagnostics) == 0 {
		// Only add this error if we haven't already added validation errors
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 100},
			},
			Severity: ptr(protocol.DiagnosticSeverityError),
			Source:   ptr("yapi"),
			Message:  fmt.Sprintf("Project config error: %v", err),
		})
	}

	// Send diagnostics
	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func findFieldLineInText(text string, fieldName string) int {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, fieldName+":") {
			return i
		}
	}
	return 0
}

func ptr[T any](v T) *T {
	return &v
}

func severityToProtocol(s validation.Severity) protocol.DiagnosticSeverity {
	switch s {
	case validation.SeverityError:
		return protocol.DiagnosticSeverityError
	case validation.SeverityWarning:
		return protocol.DiagnosticSeverityWarning
	case validation.SeverityInfo:
		return protocol.DiagnosticSeverityInformation
	default:
		return protocol.DiagnosticSeverityInformation
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// valDesc represents a value with its description for completions
type valDesc struct {
	val  string
	desc string
}

func toValueCompletion(v valDesc) protocol.CompletionItem {
	return protocol.CompletionItem{
		Label:         v.val,
		Kind:          ptr(protocol.CompletionItemKindValue),
		Detail:        ptr(v.desc),
		InsertText:    ptr(v.val),
		Documentation: v.desc,
	}
}

// Schema definitions for completions
var topLevelKeys = []struct {
	key  string
	desc string
}{
	{"url", "The target URL (required)"},
	{"path", "URL path to append"},
	{"method", "HTTP method or protocol (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, grpc, tcp)"},
	{"headers", "HTTP headers as key-value pairs"},
	{"content_type", "Content-Type header value"},
	{"body", "Request body as key-value pairs"},
	{"json", "Raw JSON string for request body"},
	{"query", "Query parameters as key-value pairs"},
	{"graphql", "GraphQL query or mutation (multiline string)"},
	{"variables", "GraphQL variables as key-value pairs"},
	{"service", "gRPC service name"},
	{"rpc", "gRPC method name"},
	{"proto", "Path to .proto file"},
	{"proto_path", "Import path for proto files"},
	{"data", "Raw data for TCP requests"},
	{"encoding", "Data encoding (text, hex, base64)"},
	{"jq_filter", "JQ filter to apply to response"},
	{"insecure", "Skip TLS verification for HTTP/GraphQL; use insecure transport for gRPC (boolean)"},
	{"plaintext", "Use plaintext gRPC (boolean)"},
	{"read_timeout", "TCP read timeout in seconds"},
	{"close_after_send", "Close TCP connection after sending (boolean)"},
	{"delay", "Wait before executing this step (e.g. 5s, 500ms)"},
}

var methodValues = []valDesc{
	{constants.MethodGET, "HTTP GET request"},
	{constants.MethodPOST, "HTTP POST request"},
	{constants.MethodPUT, "HTTP PUT request"},
	{constants.MethodDELETE, "HTTP DELETE request"},
	{constants.MethodPATCH, "HTTP PATCH request"},
	{constants.MethodHEAD, "HTTP HEAD request"},
	{constants.MethodOPTIONS, "HTTP OPTIONS request"},
}

var encodingValues = []valDesc{
	{"text", "Plain text encoding"},
	{"hex", "Hexadecimal encoding"},
	{"base64", "Base64 encoding"},
}

var contentTypeValues = []valDesc{
	{"application/json", "JSON content type"},
}

func textDocumentCompletion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	uri := params.TextDocument.URI
	doc, ok := docs[uri]
	if !ok {
		return nil, nil
	}

	line := params.Position.Line
	char := params.Position.Character

	lines := strings.Split(doc.Text, "\n")
	if int(line) >= len(lines) {
		return nil, nil
	}

	currentLine := lines[line]
	textBeforeCursor := ""
	if int(char) <= len(currentLine) {
		textBeforeCursor = currentLine[:char]
	}

	var items []protocol.CompletionItem

	// Check if we're completing a value (after a colon)
	if colonIdx := strings.Index(textBeforeCursor, ":"); colonIdx != -1 {
		key := strings.TrimSpace(textBeforeCursor[:colonIdx])

		switch key {
		case "method":
			items = utils.Map(methodValues, toValueCompletion)
		case "encoding":
			items = utils.Map(encodingValues, toValueCompletion)
		case "content_type":
			items = utils.Map(contentTypeValues, toValueCompletion)
		case "insecure", "plaintext", "close_after_send":
			items = append(items,
				protocol.CompletionItem{
					Label:      "true",
					Kind:       ptr(protocol.CompletionItemKindValue),
					InsertText: ptr("true"),
				},
				protocol.CompletionItem{
					Label:      "false",
					Kind:       ptr(protocol.CompletionItemKindValue),
					InsertText: ptr("false"),
				},
			)
		}
	} else {
		// Completing a key at the start of a line
		trimmed := strings.TrimSpace(textBeforeCursor)

		// Find which keys are already used
		usedKeys := make(map[string]bool)
		for _, l := range lines {
			if colonIdx := strings.Index(l, ":"); colonIdx != -1 {
				k := strings.TrimSpace(l[:colonIdx])
				usedKeys[k] = true
			}
		}

		for _, k := range topLevelKeys {
			if usedKeys[k.key] {
				continue
			}
			// Filter by what user has typed
			if trimmed != "" && !strings.HasPrefix(k.key, trimmed) {
				continue
			}
			items = append(items, protocol.CompletionItem{
				Label:         k.key,
				Kind:          ptr(protocol.CompletionItemKindField),
				Detail:        ptr(k.desc),
				InsertText:    ptr(k.key + ": "),
				Documentation: k.desc,
			})
		}
	}

	return items, nil
}

func textDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := params.TextDocument.URI
	doc, ok := docs[uri]
	if !ok {
		return nil, nil
	}

	line := int(params.Position.Line)
	char := int(params.Position.Character)

	// Find all env var references in the document
	refs := validation.FindEnvVarRefs(doc.Text)

	// Check if cursor is within any env var reference
	for _, ref := range refs {
		if ref.Line == line && char >= ref.Col && char <= ref.EndIndex {
			var content string
			if ref.IsDefined {
				redacted := validation.RedactValue(ref.Value)
				content = fmt.Sprintf("**Environment Variable: `%s`**\n\nValue: `%s`", ref.Name, redacted)
			} else {
				content = fmt.Sprintf("**Environment Variable: `%s`**\n\n_Not defined_", ref.Name)
			}

			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: content,
				},
				Range: &protocol.Range{
					Start: protocol.Position{Line: protocol.UInteger(line), Character: protocol.UInteger(ref.Col)},
					End:   protocol.Position{Line: protocol.UInteger(line), Character: protocol.UInteger(ref.EndIndex)},
				},
			}, nil
		}
	}

	return nil, nil
}

func textDocumentDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	uri := params.TextDocument.URI
	doc, ok := docs[uri]
	if !ok {
		return nil, nil
	}

	// No project context - can't find definitions
	if doc.Project == nil {
		return nil, nil
	}

	line := int(params.Position.Line)
	char := int(params.Position.Character)

	// Find all env var references in the document
	refs := validation.FindEnvVarRefs(doc.Text)

	// Check if cursor is within any env var reference
	for _, ref := range refs {
		if ref.Line == line && char >= ref.Col && char <= ref.EndIndex {
			// Skip chain references (e.g., ${step.field})
			if strings.Contains(ref.Name, ".") {
				return nil, nil
			}

			// Find where this variable is defined
			location, err := findVariableDefinition(ref.Name, doc.Project, doc.ProjectRoot)
			if err != nil {
				return nil, nil
			}

			return location, nil
		}
	}

	return nil, nil
}

// findVariableDefinition locates where a variable is defined
// Returns location in yapi.config.yml or .env file
func findVariableDefinition(varName string, project *config.ProjectConfigV1, projectRoot string) (*protocol.Location, error) {
	// Determine which environment to use
	envName := getEffectiveEnvironment(project)

	// Resolve environment variables to check where the variable is defined
	envVars, err := project.ResolveEnvFiles(projectRoot, envName)
	if err != nil {
		return nil, err
	}

	// Check if variable exists at all
	_, varExists := envVars[varName]
	if !varExists {
		// Variable not defined anywhere
		return nil, nil
	}

	// Check if variable is defined in yapi.config.yml
	// First check the specific environment's vars section
	if env, ok := project.Environments[envName]; ok {
		if _, inEnvVars := env.Vars[varName]; inEnvVars {
			// Variable is in environments.[envName].vars
			return findVarPositionInYAML(projectRoot, varName, []string{"environments", envName, "vars"})
		}
	}

	// Check defaults.vars
	if _, inDefaultVars := project.Defaults.Vars[varName]; inDefaultVars {
		return findVarPositionInYAML(projectRoot, varName, []string{"defaults", "vars"})
	}

	// Variable must be in .env file - try to find it there
	if env, ok := project.Environments[envName]; ok {
		for _, envFile := range env.EnvFiles {
			location, err := findVarPositionInEnvFile(projectRoot, envFile, varName)
			if err == nil && location != nil {
				return location, nil
			}
		}
	}

	// Also check defaults.env_files
	for _, envFile := range project.Defaults.EnvFiles {
		location, err := findVarPositionInEnvFile(projectRoot, envFile, varName)
		if err == nil && location != nil {
			return location, nil
		}
	}

	// Variable exists but we couldn't find its definition (might be in OS env)
	return nil, nil
}

// findVarPositionInYAML finds the position of a variable in yapi.config.yml
func findVarPositionInYAML(projectRoot string, varName string, section []string) (*protocol.Location, error) {
	// Try both .yml and .yaml extensions
	var configPath string
	ymlPath := filepath.Join(projectRoot, "yapi.config.yml")
	yamlPath := filepath.Join(projectRoot, "yapi.config.yaml")

	if _, err := os.Stat(ymlPath); err == nil {
		configPath = ymlPath
	} else if _, err := os.Stat(yamlPath); err == nil {
		configPath = yamlPath
	} else {
		return nil, fmt.Errorf("config file not found")
	}

	// Read and parse the YAML file
	contentBytes, err := os.ReadFile(configPath) // #nosec G304 -- configPath is constructed from validated projectRoot
	if err != nil {
		return nil, err
	}
	content := string(contentBytes)

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return nil, err
	}

	// Navigate to the section (e.g., ["environments", "dev", "vars"])
	currentNode := &root
	if len(root.Content) > 0 {
		currentNode = root.Content[0] // Get the document content
	}

	for _, key := range section {
		valueNode := findNodeInMapping(currentNode, key)
		if valueNode == nil {
			return nil, fmt.Errorf("section not found: %s", key)
		}
		currentNode = valueNode
	}

	// Now find the key node for the variable
	keyNode := findKeyNodeInMapping(currentNode, varName)
	if keyNode == nil {
		return nil, fmt.Errorf("variable not found in section")
	}

	// Convert YAML position (1-indexed) to LSP position (0-indexed)
	startPos := protocol.Position{
		Line:      protocol.UInteger(keyNode.Line - 1),
		Character: protocol.UInteger(keyNode.Column - 1),
	}
	endPos := protocol.Position{
		Line:      protocol.UInteger(keyNode.Line - 1),
		Character: protocol.UInteger(keyNode.Column - 1 + len(varName)),
	}

	return &protocol.Location{
		URI: protocol.DocumentUri("file://" + configPath),
		Range: protocol.Range{
			Start: startPos,
			End:   endPos,
		},
	}, nil
}

// findVarPositionInEnvFile finds the position of a variable in an .env file
func findVarPositionInEnvFile(projectRoot string, envFile string, varName string) (*protocol.Location, error) {
	envPath := filepath.Join(projectRoot, envFile)
	if _, err := os.Stat(envPath); err != nil {
		return nil, fmt.Errorf("env file not found: %s", envFile)
	}

	contentBytes, err := os.ReadFile(envPath) // #nosec G304 -- envPath is constructed from validated projectRoot and envFile
	if err != nil {
		return nil, err
	}
	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 1 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == varName {
			// Found the variable - calculate its position
			col := strings.Index(line, key)
			startPos := protocol.Position{
				Line:      protocol.UInteger(lineNum),
				Character: protocol.UInteger(col),
			}
			endPos := protocol.Position{
				Line:      protocol.UInteger(lineNum),
				Character: protocol.UInteger(col + len(key)),
			}

			return &protocol.Location{
				URI: protocol.DocumentUri("file://" + envPath),
				Range: protocol.Range{
					Start: startPos,
					End:   endPos,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("variable not found in env file")
}

// findNodeInMapping finds the value node for a given key in a YAML mapping
func findNodeInMapping(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	// MappingNode content is [key, value, key, value, ...]
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 < len(node.Content) && node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

// findKeyNodeInMapping finds the key node itself (not the value) in a YAML mapping
func findKeyNodeInMapping(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	// MappingNode content is [key, value, key, value, ...]
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}

	return nil
}

// getEffectiveEnvironment returns the environment name to use for lookups
func getEffectiveEnvironment(project *config.ProjectConfigV1) string {
	// Use default_environment if set
	if project.DefaultEnvironment != "" {
		return project.DefaultEnvironment
	}

	// Otherwise use the first environment alphabetically
	if len(project.Environments) > 0 {
		var firstEnv string
		for envName := range project.Environments {
			if firstEnv == "" || envName < firstEnv {
				firstEnv = envName
			}
		}
		return firstEnv
	}

	return ""
}

// uriToPath converts a file:// URI to a filesystem path.
func uriToPath(uri protocol.DocumentUri) string {
	u, err := url.Parse(string(uri))
	if err != nil {
		return ""
	}
	if u.Scheme != "file" {
		return ""
	}
	return u.Path
}

package langserver

import (
	"fmt"
	"strings"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"yapi.run/cli/internal/compiler"
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
	URI  protocol.DocumentUri
	Text string
}

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
	}

	srv := server.NewServer(&handler, lsName, false)
	srv.RunStdio()
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

	docs[uri] = &document{
		URI:  uri,
		Text: text,
	}

	validateAndNotify(ctx, uri, text)
	return nil
}

func textDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI

	// With TextDocumentSyncKindFull, we get the full text in each change
	if len(params.ContentChanges) > 0 {
		text := params.ContentChanges[len(params.ContentChanges)-1].(protocol.TextDocumentContentChangeEventWhole).Text

		if doc, ok := docs[uri]; ok {
			doc.Text = text
		} else {
			docs[uri] = &document{
				URI:  uri,
				Text: text,
			}
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
	analysis, err := validation.AnalyzeConfigString(text)
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
		var resolver vars.Resolver = MockResolver
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

// MockResolver provides placeholder values for variable interpolation in LSP validation.
// This allows the compiler to validate the config structure even without real env vars.
func MockResolver(key string) (string, error) {
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "port") {
		return "8080", nil
	}
	if strings.Contains(keyLower, "host") {
		return "localhost", nil
	}
	if strings.Contains(keyLower, "url") {
		return "http://localhost:8080", nil
	}
	// Return a placeholder for anything else - don't fail
	return "PLACEHOLDER", nil
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
	{"insecure", "Skip TLS verification (boolean)"},
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

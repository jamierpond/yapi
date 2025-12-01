package langserver

import (
	"fmt"
	"strings"
	"yapi.run/internal/config"
	"yapi.run/internal/envsubst"
	"yapi.run/internal/validation"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"gopkg.in/yaml.v3"
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
	diagnostics := []protocol.Diagnostic{}

	var cfg config.YapiConfig
	if err := yaml.Unmarshal([]byte(text), &cfg); err != nil {
		// YAML parse error - show at line 0
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 1},
			},
			Severity: ptr(protocol.DiagnosticSeverityError),
			Source:   ptr("yapi"),
			Message:  "invalid YAML: " + err.Error(),
		})
	} else {
		issues := validation.ValidateConfig(&cfg)
		for _, issue := range issues {
			line := findFieldLine(text, issue.Field)
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: line, Character: 0},
					End:   protocol.Position{Line: line, Character: 100},
				},
				Severity: ptr(severityToProtocol(issue.Severity)),
				Source:   ptr("yapi"),
				Message:  issue.Message,
			})
		}

		// GraphQL syntax validation
		if cfg.Graphql != "" {
			gqlDiags := validateGraphQLSyntax(text, cfg.Graphql)
			diagnostics = append(diagnostics, gqlDiags...)
		}

		// Environment variable validation
		envDiags := validateEnvVars(text)
		diagnostics = append(diagnostics, envDiags...)
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func findFieldLine(text string, field string) protocol.UInteger {
	if field == "" {
		return 0
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), field+":") {
			return protocol.UInteger(i)
		}
	}
	return 0
}

func validateGraphQLSyntax(fullYamlText string, gqlQuery string) []protocol.Diagnostic {
	src := source.NewSource(&source.Source{
		Body: []byte(gqlQuery),
		Name: "GraphQL Query",
	})

	_, err := parser.Parse(parser.ParseParams{Source: src})
	if err == nil {
		return nil
	}

	// Find where the "graphql:" block starts in the YAML file
	blockStartLine := findFieldLine(fullYamlText, "graphql")

	// The query content starts on the line after "graphql: |"
	// so we add 1 to get to the actual query content
	targetLine := blockStartLine + 1

	return []protocol.Diagnostic{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: targetLine, Character: 0},
				End:   protocol.Position{Line: targetLine + 1, Character: 0},
			},
			Severity: ptr(protocol.DiagnosticSeverityError),
			Source:   ptr("yapi"),
			Message:  "GraphQL syntax error: " + err.Error(),
		},
	}
}

func validateEnvVars(text string) []protocol.Diagnostic {
	var diagnostics []protocol.Diagnostic
	lines := strings.Split(text, "\n")

	for lineNum, line := range lines {
		refs := envsubst.FindAllWithPositions(line)
		for _, ref := range refs {
			missing := envsubst.FindMissing(line[ref.Start:ref.End])
			if len(missing) > 0 {
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      protocol.UInteger(lineNum),
							Character: protocol.UInteger(ref.Start),
						},
						End: protocol.Position{
							Line:      protocol.UInteger(lineNum),
							Character: protocol.UInteger(ref.End),
						},
					},
					Severity: ptr(protocol.DiagnosticSeverityWarning),
					Source:   ptr("yapi"),
					Message:  fmt.Sprintf("environment variable %q is not set", ref.Name),
				})
			}
		}
	}

	return diagnostics
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

func severityToMessageType(s validation.Severity) protocol.MessageType {
	switch s {
	case validation.SeverityError:
		return protocol.MessageTypeError
	case validation.SeverityWarning:
		return protocol.MessageTypeWarning
	case validation.SeverityInfo:
		return protocol.MessageTypeInfo
	default:
		return protocol.MessageTypeInfo
	}
}

func boolPtr(b bool) *bool {
	return &b
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
}

var methodValues = []struct {
	val  string
	desc string
}{
	{"GET", "HTTP GET request"},
	{"POST", "HTTP POST request"},
	{"PUT", "HTTP PUT request"},
	{"DELETE", "HTTP DELETE request"},
	{"PATCH", "HTTP PATCH request"},
	{"HEAD", "HTTP HEAD request"},
	{"OPTIONS", "HTTP OPTIONS request"},
	{"grpc", "gRPC request (deprecated)"},
	{"tcp", "Raw TCP request (deprecated)"},
}

var encodingValues = []struct {
	val  string
	desc string
}{
	{"text", "Plain text encoding"},
	{"hex", "Hexadecimal encoding"},
	{"base64", "Base64 encoding"},
}

var contentTypeValues = []struct {
	val  string
	desc string
}{
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
			for _, m := range methodValues {
				items = append(items, protocol.CompletionItem{
					Label:         m.val,
					Kind:          ptr(protocol.CompletionItemKindValue),
					Detail:        ptr(m.desc),
					InsertText:    ptr(m.val),
					Documentation: m.desc,
				})
			}
		case "encoding":
			for _, e := range encodingValues {
				items = append(items, protocol.CompletionItem{
					Label:         e.val,
					Kind:          ptr(protocol.CompletionItemKindValue),
					Detail:        ptr(e.desc),
					InsertText:    ptr(e.val),
					Documentation: e.desc,
				})
			}
		case "content_type":
			for _, ct := range contentTypeValues {
				items = append(items, protocol.CompletionItem{
					Label:         ct.val,
					Kind:          ptr(protocol.CompletionItemKindValue),
					Detail:        ptr(ct.desc),
					InsertText:    ptr(ct.val),
					Documentation: ct.desc,
				})
			}
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

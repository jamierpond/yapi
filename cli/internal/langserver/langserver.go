package langserver

import (
	"cli/internal/config"
	"cli/internal/validation"
	"strings"

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

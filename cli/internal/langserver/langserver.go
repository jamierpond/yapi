package langserver

import (
	"cli/internal/config"
	"cli/internal/validation"

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
	var cfg config.YapiConfig
	if err := yaml.Unmarshal([]byte(text), &cfg); err != nil {
		ctx.Notify(protocol.ServerWindowShowMessage, protocol.ShowMessageParams{
			Type:    protocol.MessageTypeError,
			Message: "yapi: invalid YAML: " + err.Error(),
		})
		return
	}

	issues := validation.ValidateConfig(&cfg)
	for _, issue := range issues {
		msgType := severityToMessageType(issue.Severity)
		ctx.Notify(protocol.ServerWindowShowMessage, protocol.ShowMessageParams{
			Type:    msgType,
			Message: "yapi: " + issue.Message,
		})
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

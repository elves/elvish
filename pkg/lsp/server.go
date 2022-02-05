package lsp

import (
	"context"
	"encoding/json"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

var (
	errMethodNotFound = &jsonrpc2.Error{
		Code: jsonrpc2.CodeMethodNotFound, Message: "method not found"}
	errInvalidParams = &jsonrpc2.Error{
		Code: jsonrpc2.CodeInvalidParams, Message: "invalid params"}
)

func routingHandler(methods map[string]method) jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		fn, ok := methods[req.Method]
		if !ok {
			return nil, errMethodNotFound
		}
		return fn(ctx, conn, *req.Params)
	})
}

type method func(context.Context, jsonrpc2.JSONRPC2, json.RawMessage) (interface{}, error)

func noop(_ context.Context, _ jsonrpc2.JSONRPC2, _ json.RawMessage) (interface{}, error) {
	return nil, nil
}

type server struct {
}

func (s *server) handler() jsonrpc2.Handler {
	return routingHandler(map[string]method{
		"initialize":             s.initialize,
		"textDocument/didOpen":   s.didOpen,
		"textDocument/didChange": s.didChange,

		"textDocument/didClose": noop,
		// Required by spec.
		"initialized": noop,
		// Called by clients even when server doesn't advertise support:
		// https://microsoft.github.io/language-server-protocol/specification#workspace_didChangeWatchedFiles
		"workspace/didChangeWatchedFiles": noop,
	})
}

func (s *server) initialize(_ context.Context, _ jsonrpc2.JSONRPC2, _ json.RawMessage) (interface{}, error) {
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
				},
			},
		},
	}, nil
}

func (s *server) didOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (interface{}, error) {
	var params lsp.DidOpenTextDocumentParams
	if json.Unmarshal(rawParams, &params) != nil {
		return nil, errInvalidParams
	}

	uri, content := params.TextDocument.URI, params.TextDocument.Text
	go update(ctx, conn, uri, content)
	return nil, nil
}

func (s *server) didChange(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (interface{}, error) {
	var params lsp.DidChangeTextDocumentParams
	if json.Unmarshal(rawParams, &params) != nil {
		return nil, errInvalidParams
	}

	uri, content := params.TextDocument.URI, params.ContentChanges[0].Text
	go update(ctx, conn, uri, content)
	return nil, nil
}

func update(ctx context.Context, conn jsonrpc2.JSONRPC2, uri lsp.DocumentURI, content string) {
	conn.Notify(ctx, "textDocument/publishDiagnostics",
		lsp.PublishDiagnosticsParams{URI: uri, Diagnostics: diagnostics(uri, content)})
}

func diagnostics(fileURI lsp.DocumentURI, content string) []lsp.Diagnostic {
	_, err := parse.Parse(parse.Source{Name: string(fileURI), Code: content}, parse.Config{})
	if err == nil {
		return []lsp.Diagnostic{}
	}

	entries := err.(*parse.Error).Entries
	diags := make([]lsp.Diagnostic, len(entries))
	for i, err := range entries {
		diags[i] = lsp.Diagnostic{
			Range:    rangeToLSP(content, err),
			Severity: lsp.Error,
			Source:   "parse",
			Message:  err.Message,
		}
	}
	return diags
}

func rangeToLSP(s string, r diag.Ranger) lsp.Range {
	rg := r.Range()
	return lsp.Range{
		Start: position(s, rg.From),
		End:   position(s, rg.To),
	}
}

func position(s string, idx int) lsp.Position {
	var pos lsp.Position
	lastCR := false

	for i, r := range s {
		if i == idx {
			return pos
		}
		switch {
		case r == '\r':
			pos.Line++
			pos.Character = 0
		case r == '\n':
			if lastCR {
				// Ignore \n if it's part of a \r\n sequence
			} else {
				pos.Line++
				pos.Character = 0
			}
		case r <= 0xFFFF:
			// Encoded in UTF-16 with one unit
			pos.Character++
		default:
			// Encoded in UTF-16 with two units
			pos.Character += 2
		}
		lastCR = r == '\r'
	}
	return pos
}

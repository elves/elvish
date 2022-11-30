package lsp

import (
	"context"
	"encoding/json"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/edit/complete"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

var (
	errMethodNotFound = &jsonrpc2.Error{
		Code: jsonrpc2.CodeMethodNotFound, Message: "method not found"}
	errInvalidParams = &jsonrpc2.Error{
		Code: jsonrpc2.CodeInvalidParams, Message: "invalid params"}
)

type server struct {
	evaler  *eval.Evaler
	content map[lsp.DocumentURI]string
}

func newServer() *server {
	return &server{eval.NewEvaler(), make(map[lsp.DocumentURI]string)}
}

func handler(s *server) jsonrpc2.Handler {
	return routingHandler(map[string]method{
		"initialize":              s.initialize,
		"textDocument/didOpen":    s.didOpen,
		"textDocument/didChange":  s.didChange,
		"textDocument/hover":      s.hover,
		"textDocument/completion": s.completion,

		"textDocument/didClose": noop,
		// Required by spec.
		"initialized": noop,
		// Called by clients even when server doesn't advertise support:
		// https://microsoft.github.io/language-server-protocol/specification#workspace_didChangeWatchedFiles
		"workspace/didChangeWatchedFiles": noop,
	})
}

type method func(context.Context, jsonrpc2.JSONRPC2, json.RawMessage) (any, error)

func noop(_ context.Context, _ jsonrpc2.JSONRPC2, _ json.RawMessage) (any, error) {
	return nil, nil
}

func routingHandler(methods map[string]method) jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		fn, ok := methods[req.Method]
		if !ok {
			return nil, errMethodNotFound
		}
		return fn(ctx, conn, *req.Params)
	})
}

// Handler implementations. These are all called synchronously.

func (s *server) initialize(_ context.Context, _ jsonrpc2.JSONRPC2, _ json.RawMessage) (any, error) {
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{},
		},
	}, nil
}

func (s *server) didOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (any, error) {
	var params lsp.DidOpenTextDocumentParams
	if json.Unmarshal(rawParams, &params) != nil {
		return nil, errInvalidParams
	}

	uri, content := params.TextDocument.URI, params.TextDocument.Text
	s.content[uri] = content
	go publishDiagnostics(ctx, conn, uri, content)
	return nil, nil
}

func (s *server) didChange(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (any, error) {
	var params lsp.DidChangeTextDocumentParams
	if json.Unmarshal(rawParams, &params) != nil {
		return nil, errInvalidParams
	}

	// ContentChanges includes full text since the server is only advertised to
	// support that; see the initialize method.
	uri, content := params.TextDocument.URI, params.ContentChanges[0].Text
	s.content[uri] = content
	go publishDiagnostics(ctx, conn, uri, content)
	return nil, nil
}

func (s *server) hover(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (any, error) {
	return lsp.Hover{}, nil
}

func (s *server) completion(ctx context.Context, conn jsonrpc2.JSONRPC2, rawParams json.RawMessage) (any, error) {
	var params lsp.CompletionParams
	if json.Unmarshal(rawParams, &params) != nil {
		return nil, errInvalidParams
	}

	content := s.content[params.TextDocument.URI]
	result, err := complete.Complete(
		complete.CodeBuffer{
			Content: content,
			Dot:     lspPositionToIdx(content, params.Position)},
		s.evaler,
		complete.Config{},
	)

	if err != nil {
		return []lsp.CompletionItem{}, nil
	}

	lspItems := make([]lsp.CompletionItem, len(result.Items))
	lspRange := lspRangeFromRange(content, result.Replace)
	var kind lsp.CompletionItemKind
	switch result.Name {
	case "command":
		kind = lsp.CIKFunction
	case "variable":
		kind = lsp.CIKVariable
	default:
		// TODO: Support more values of kind
	}
	for i, item := range result.Items {
		lspItems[i] = lsp.CompletionItem{
			Label: item.ToInsert,
			Kind:  kind,
			TextEdit: &lsp.TextEdit{
				Range:   lspRange,
				NewText: item.ToInsert,
			},
		}
	}
	return lspItems, nil
}

func publishDiagnostics(ctx context.Context, conn jsonrpc2.JSONRPC2, uri lsp.DocumentURI, content string) {
	conn.Notify(ctx, "textDocument/publishDiagnostics",
		lsp.PublishDiagnosticsParams{URI: uri, Diagnostics: diagnostics(uri, content)})
}

func diagnostics(uri lsp.DocumentURI, content string) []lsp.Diagnostic {
	_, err := parse.Parse(parse.Source{Name: string(uri), Code: content}, parse.Config{})
	if err == nil {
		return []lsp.Diagnostic{}
	}

	entries := parse.UnpackErrors(err)
	diags := make([]lsp.Diagnostic, len(entries))
	for i, err := range entries {
		diags[i] = lsp.Diagnostic{
			Range:    lspRangeFromRange(content, err),
			Severity: lsp.Error,
			Source:   "parse",
			Message:  err.Message,
		}
	}
	return diags
}

func lspRangeFromRange(s string, r diag.Ranger) lsp.Range {
	rg := r.Range()
	return lsp.Range{
		Start: lspPositionFromIdx(s, rg.From),
		End:   lspPositionFromIdx(s, rg.To),
	}
}

func lspPositionToIdx(s string, pos lsp.Position) int {
	var idx int
	walkString(s, func(i int, p lsp.Position) bool {
		idx = i
		return p.Line < pos.Line || (p.Line == pos.Line && p.Character < pos.Character)
	})
	return idx
}

func lspPositionFromIdx(s string, idx int) lsp.Position {
	var pos lsp.Position
	walkString(s, func(i int, p lsp.Position) bool {
		pos = p
		return i < idx
	})
	return pos
}

// Generates (index, lspPosition) pairs in s, stopping if f returns false.
func walkString(s string, f func(i int, p lsp.Position) bool) {
	var p lsp.Position
	lastCR := false

	for i, r := range s {
		if !f(i, p) {
			return
		}
		switch {
		case r == '\r':
			p.Line++
			p.Character = 0
		case r == '\n':
			if lastCR {
				// Ignore \n if it's part of a \r\n sequence
			} else {
				p.Line++
				p.Character = 0
			}
		case r <= 0xFFFF:
			// Encoded in UTF-16 with one unit
			p.Character++
		default:
			// Encoded in UTF-16 with two units
			p.Character += 2
		}
		lastCR = r == '\r'
	}
	f(len(s), p)
}

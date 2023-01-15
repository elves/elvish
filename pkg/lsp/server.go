package lsp

import (
	"context"
	"encoding/json"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/edit/complete"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
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
		"textDocument/didOpen":    convertMethod(s.didOpen),
		"textDocument/didChange":  convertMethod(s.didChange),
		"textDocument/hover":      s.hover,
		"textDocument/completion": convertMethod(s.completion),

		"textDocument/didClose": noop,
		// Required by spec.
		"initialized": noop,
		// Called by clients even when server doesn't advertise support:
		// https://microsoft.github.io/language-server-protocol/specification#workspace_didChangeWatchedFiles
		"workspace/didChangeWatchedFiles": noop,

		"elvish/eval": convertMethod(s.eval),
	})
}

type method func(context.Context, json.RawMessage) (any, error)

func convertMethod[T any](f func(context.Context, T) (any, error)) method {
	return func(ctx context.Context, rawParams json.RawMessage) (any, error) {
		var params T
		if json.Unmarshal(rawParams, &params) != nil {
			return nil, errInvalidParams
		}
		return f(ctx, params)
	}
}

func noop(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil }

type connKey struct{}

func routingHandler(methods map[string]method) jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		fn, ok := methods[req.Method]
		if !ok {
			return nil, errMethodNotFound
		}
		return fn(context.WithValue(ctx, connKey{}, conn), *req.Params)
	})
}

func conn(ctx context.Context) *jsonrpc2.Conn { return ctx.Value(connKey{}).(*jsonrpc2.Conn) }

// Handler implementations. These are all called synchronously.

func (s *server) initialize(_ context.Context, _ json.RawMessage) (any, error) {
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

func (s *server) didOpen(ctx context.Context, params lsp.DidOpenTextDocumentParams) (any, error) {
	uri, content := params.TextDocument.URI, params.TextDocument.Text
	s.content[uri] = content
	go publishDiagnostics(ctx, uri, content)
	return nil, nil
}

func (s *server) didChange(ctx context.Context, params lsp.DidChangeTextDocumentParams) (any, error) {
	// ContentChanges includes full text since the server is only advertised to
	// support that; see the initialize method.
	uri, content := params.TextDocument.URI, params.ContentChanges[0].Text
	s.content[uri] = content
	go publishDiagnostics(ctx, uri, content)
	return nil, nil
}

func (s *server) hover(_ context.Context, rawParams json.RawMessage) (any, error) {
	return lsp.Hover{}, nil
}

func (s *server) completion(_ context.Context, params lsp.CompletionParams) (any, error) {
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

// evalParams and evalResults are modelled after nREPL
// (https://nrepl.org/nrepl/1.0/design/overview.html).

type evalParams struct {
	Code string `json:"code"`
}

type evalResults struct {
	Out   string   `json:"out"`
	Err   string   `json:"err"`
	Value []string `json:"value"`
}

func (s *server) eval(_ context.Context, params evalParams) (any, error) {
	port1, collect1, err := eval.CapturePort()
	if err != nil {
		return nil, err
	}
	port2, collect2, err := eval.CapturePort()
	if err != nil {
		return nil, err
	}
	ports := []*eval.Port{eval.DummyInputPort, port1, port2}

	err = s.evaler.Eval(parse.Source{Name: "[lsp]", Code: params.Code}, eval.EvalCfg{Ports: ports})
	// TODO: Send error too
	_ = err

	values, stdout := collect1()
	valueStrings := make([]string, len(values))
	for i, value := range values {
		valueStrings[i] = vals.ToString(value)
	}
	_, stderr := collect2()
	return evalResults{string(stdout), string(stderr), valueStrings}, nil
}

func publishDiagnostics(ctx context.Context, uri lsp.DocumentURI, content string) {
	conn(ctx).Notify(ctx, "textDocument/publishDiagnostics",
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

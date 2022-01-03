package main

import (
	"context"
	"fmt"
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

type filecontext struct {
}

type server struct {
	rootURI      string
	files        map[string]string
	fileContexts map[string]filecontext
}

func (s *server) Initialize(ctx context.Context, conn jsonrpc2.JSONRPC2, params lsp.InitializeParams) (*lsp.InitializeResult, *lsp.InitializeError) {
	s.rootURI = string(params.RootURI)
	s.files = map[string]string{}
	s.fileContexts = map[string]filecontext{}

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

func (s *server) Initialized(ctx context.Context, conn jsonrpc2.JSONRPC2, params struct{}) {
	// we don't need to do anything, we just need to have this
	// here
}

func (s *server) DidOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, params lsp.DidOpenTextDocumentParams) {
	s.files[strings.TrimPrefix(string(params.TextDocument.URI), s.rootURI)] = params.TextDocument.Text
	go s.evaluate(ctx, conn, params.TextDocument.URI, params.TextDocument.Text)
}

func (s *server) DidChange(ctx context.Context, conn jsonrpc2.JSONRPC2, params lsp.DidChangeTextDocumentParams) {
	s.files[strings.TrimPrefix(string(params.TextDocument.URI), s.rootURI)] = params.ContentChanges[0].Text
	go s.evaluate(ctx, conn, params.TextDocument.URI, params.ContentChanges[0].Text)
}

func (s *server) DidChangeWatchedFiles(ctx context.Context, conn jsonrpc2.JSONRPC2, params lsp.DidChangeWatchedFilesParams) {
	// we don't currently need this, vscode just complains
	// when it's not available
}

func (s *server) DidClose(ctx context.Context, conn jsonrpc2.JSONRPC2, params lsp.DidCloseTextDocumentParams) {
	delete(s.files, strings.TrimPrefix(string(params.TextDocument.URI), s.rootURI))
	delete(s.fileContexts, strings.TrimPrefix(string(params.TextDocument.URI), s.rootURI))
}

// TODO: account for multi-byte characters
func idxToPos(str string, idx int) lsp.Position {
	col := 0
	line := 0

	for i, c := range str {
		if c == '\n' {
			col = 0
			line++
		} else {
			col++
		}

		if i == idx {
			return lsp.Position{Line: line, Character: col}
		}
	}

	panic(fmt.Sprintf("out of range: wanted %d, only had length %d", idx, len(str)))
}

func rangeToLSP(s string, r diag.Ranging) lsp.Range {
	return lsp.Range{
		Start: idxToPos(s, r.From),
		End:   idxToPos(s, r.To),
	}
}

func rangerToLSP(s string, r diag.Ranger) lsp.Range {
	return rangeToLSP(s, r.Range())
}

func (s *server) collectParseDiagnostics(ctx context.Context, fileURI string, diags *lsp.PublishDiagnosticsParams) {
	content := s.files[fileURI]

	_, err := parse.Parse(parse.SourceForTest(content), parse.Config{WarningWriter: nil})
	if err == nil {
		return
	}
	var (
		v  *parse.Error
		ok bool
	)
	if v, ok = err.(*parse.Error); !ok {
		return
	}

	for _, err := range v.Entries {
		diags.Diagnostics = append(diags.Diagnostics, lsp.Diagnostic{
			Range:    rangerToLSP(content, err),
			Severity: lsp.Error,
			Source:   "parse",
			Message:  err.Message,
		})
	}
}

func (s *server) evaluate(ctx context.Context, conn jsonrpc2.JSONRPC2, uri lsp.DocumentURI, content string) {
	fileURI := strings.TrimPrefix(string(uri), s.rootURI)

	diags := lsp.PublishDiagnosticsParams{
		URI: uri,
	}

	s.collectParseDiagnostics(ctx, fileURI, &diags)

	conn.Notify(ctx, "textDocument/publishDiagnostics", diags)
}

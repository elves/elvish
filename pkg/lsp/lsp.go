// Package lsp implements a language server for Elvish.
package lsp

import (
	"context"
	"os"

	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/prog"
)

// Program is the LSP subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) Run(fds [3]*os.File, f *prog.Flags, _ []string) error {
	if !f.LSP {
		return prog.ErrNotSuitable
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := server{}
	conn := jsonrpc2.NewConn(ctx,
		jsonrpc2.NewBufferedStream(transport{fds[0], fds[1]}, jsonrpc2.VSCodeObjectCodec{}),
		s.handler())
	<-conn.DisconnectNotify()
	return nil
}

type transport struct{ in, out *os.File }

func (c transport) Read(p []byte) (int, error)  { return c.in.Read(p) }
func (c transport) Write(p []byte) (int, error) { return c.out.Write(p) }

func (c transport) Close() error {
	if err := c.in.Close(); err != nil {
		c.out.Close()
		return err
	}
	return c.out.Close()
}

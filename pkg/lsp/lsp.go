// Package lsp implements a language server for Elvish.
package lsp

import (
	"context"
	"os"

	"github.com/sourcegraph/jsonrpc2"
	"src.elv.sh/pkg/prog"
)

// Program is the LSP subprogram.
type Program struct {
	run bool
}

func (p *Program) RegisterFlags(fs *prog.FlagSet) {
	fs.BoolVar(&p.run, "lsp", false, "run language server instead of shell")
}

func (p *Program) Run(fds [3]*os.File, _ []string) error {
	if !p.run {
		return prog.ErrNextProgram
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

package main

import (
	"context"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn := jsonrpc2.NewConn(ctx,
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		makeHandler())
	<-conn.DisconnectNotify()
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error)  { return os.Stdin.Read(p) }
func (stdrwc) Write(p []byte) (int, error) { return os.Stdout.Write(p) }

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		os.Stdout.Close()
		return err
	}
	return os.Stdout.Close()
}

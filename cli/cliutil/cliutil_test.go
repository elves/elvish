package cliutil

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
)

func TestSetGetCodeBuffer(t *testing.T) {
	app := cli.NewApp(cli.AppSpec{})
	buf := codearea.CodeBuffer{Content: "test code", Dot: 2}
	SetCodeBuffer(app, buf)
	if gotBuf := GetCodeBuffer(app); gotBuf != buf {
		t.Errorf("Got buffer %v, want %v", gotBuf, buf)
	}
}

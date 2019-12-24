package edit

import (
	"testing"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/store/storedefs"
)

func TestCommandHistory(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo 1")
		s.AddCmd("echo 2")
	}))
	defer f.Cleanup()

	// TODO(xiaq): Test session history too.
	evals(f.Evaler, `@cmds = (edit:command-history)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			vals.MakeMap("id", "0", "cmd", "echo 1"),
			vals.MakeMap("id", "1", "cmd", "echo 2")))
}

func TestInsertLastWord(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo foo bar")
	}))
	defer f.Cleanup()

	evals(f.Evaler, "edit:insert-last-word")
	wantBuf := codearea.CodeBuffer{Content: "bar", Dot: 3}
	if buf := cli.CodeBuffer(f.Editor.app); buf != wantBuf {
		t.Errorf("buf = %v, want %v", buf, wantBuf)
	}
}

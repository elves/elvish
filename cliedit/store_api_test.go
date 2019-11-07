package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/store"
)

func TestCommandHistory(t *testing.T) {
	st, cleanupStore := store.MustGetTempStore()
	defer cleanupStore()
	st.AddCmd("echo 1")
	st.AddCmd("echo 2")
	_, _, ev, cleanup := setupWithStore(st)
	defer cleanup()

	// TODO(xiaq): Test session history too.
	evalf(ev, `@cmds = (edit:command-history)`)
	wantCmds := vals.MakeList(
		vals.MakeMap("id", "0", "cmd", "echo 1"),
		vals.MakeMap("id", "1", "cmd", "echo 2"))
	if cmds := ev.Global["cmds"].Get().(vals.List); !vals.Equal(cmds, wantCmds) {
		t.Errorf("got $cmd = %v, want %v",
			vals.Repr(cmds, vals.NoPretty), vals.Repr(wantCmds, vals.NoPretty))
	}
}

func TestInsertLastWord(t *testing.T) {
	st, cleanupStore := store.MustGetTempStore()
	defer cleanupStore()
	st.AddCmd("echo foo bar")
	ed, _, ev, cleanup := setupWithStore(st)
	defer cleanup()

	evalf(ev, "edit:insert-last-word")
	wantBuf := codearea.Buffer{Content: "bar", Dot: 3}
	if buf := cliutil.GetCodeBuffer(ed.app); buf != wantBuf {
		t.Errorf("buf = %v, want %v", buf, wantBuf)
	}
}

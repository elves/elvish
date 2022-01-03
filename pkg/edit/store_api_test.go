package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/store/storedefs"
)

func TestCommandHistory(t *testing.T) {
	f := setup(t, storeOp(func(s storedefs.Store) {
		s.AddCmd("echo 0")
		s.AddCmd("echo 1")
		s.AddCmd("echo 2")
		s.AddCmd("echo 2")
		s.AddCmd("echo 1")
		s.AddCmd("echo 3")
		s.AddCmd("echo 1")
	}))

	// TODO(xiaq): Test session history too. See NewHybridStore and NewMemStore.

	evals(f.Evaler, `var @cmds = (edit:command-history)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			cmdMap(1, "echo 0"),
			cmdMap(2, "echo 1"),
			cmdMap(3, "echo 2"),
			cmdMap(4, "echo 2"),
			cmdMap(5, "echo 1"),
			cmdMap(6, "echo 3"),
			cmdMap(7, "echo 1"),
		))

	evals(f.Evaler, `var @cmds = (edit:command-history &newest-first)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			cmdMap(7, "echo 1"),
			cmdMap(6, "echo 3"),
			cmdMap(5, "echo 1"),
			cmdMap(4, "echo 2"),
			cmdMap(3, "echo 2"),
			cmdMap(2, "echo 1"),
			cmdMap(1, "echo 0"),
		))

	evals(f.Evaler, `var @cmds = (edit:command-history &dedup)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			cmdMap(1, "echo 0"),
			cmdMap(4, "echo 2"),
			cmdMap(6, "echo 3"),
			cmdMap(7, "echo 1"),
		))

	evals(f.Evaler, `var @cmds = (edit:command-history &dedup &newest-first)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			cmdMap(7, "echo 1"),
			cmdMap(6, "echo 3"),
			cmdMap(4, "echo 2"),
			cmdMap(1, "echo 0"),
		))

	evals(f.Evaler, `var @cmds = (edit:command-history &dedup &newest-first &cmd-only)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			"echo 1",
			"echo 3",
			"echo 2",
			"echo 0",
		))

	testThatOutputErrorIsBubbled(t, f, "edit:command-history")
	testThatOutputErrorIsBubbled(t, f, "edit:command-history &cmd-only")
}

func cmdMap(id int, cmd string) vals.Map {
	return vals.MakeMap("id", id, "cmd", cmd)
}

func TestInsertLastWord(t *testing.T) {
	f := setup(t, storeOp(func(s storedefs.Store) {
		s.AddCmd("echo foo bar")
	}))

	evals(f.Evaler, "edit:insert-last-word")
	wantBuf := tk.CodeBuffer{Content: "bar", Dot: 3}
	if buf := codeArea(f.Editor.app).CopyState().Buffer; buf != wantBuf {
		t.Errorf("buf = %v, want %v", buf, wantBuf)
	}
}

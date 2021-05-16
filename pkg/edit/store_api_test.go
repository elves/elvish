package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/store"
)

func TestCommandHistory(t *testing.T) {
	f := setup(storeOp(func(s store.Store) {
		s.AddCmd("echo 1")
		s.AddCmd("echo 2")
		s.AddCmd("echo 2")
		s.AddCmd("echo 1")
		s.AddCmd("echo 3")
		s.AddCmd("echo 1")
	}))
	defer f.Cleanup()

	// TODO(xiaq): Test session history too. See NewHybridStore and NewMemStore.

	evals(f.Evaler, `@cmds = (edit:command-history)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			vals.MakeMap("id", "1", "cmd", "echo 1"),
			vals.MakeMap("id", "2", "cmd", "echo 2"),
			vals.MakeMap("id", "3", "cmd", "echo 2"),
			vals.MakeMap("id", "4", "cmd", "echo 1"),
			vals.MakeMap("id", "5", "cmd", "echo 3"),
			vals.MakeMap("id", "6", "cmd", "echo 1"),
		))

	evals(f.Evaler, `@cmds = (edit:command-history &newest-first)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			vals.MakeMap("id", "6", "cmd", "echo 1"),
			vals.MakeMap("id", "5", "cmd", "echo 3"),
			vals.MakeMap("id", "4", "cmd", "echo 1"),
			vals.MakeMap("id", "3", "cmd", "echo 2"),
			vals.MakeMap("id", "2", "cmd", "echo 2"),
			vals.MakeMap("id", "1", "cmd", "echo 1"),
		))

	evals(f.Evaler, `@cmds = (edit:command-history &dedup)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			vals.MakeMap("id", "3", "cmd", "echo 2"),
			vals.MakeMap("id", "5", "cmd", "echo 3"),
			vals.MakeMap("id", "6", "cmd", "echo 1"),
		))

	evals(f.Evaler, `@cmds = (edit:command-history &dedup &newest-first)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			vals.MakeMap("id", "6", "cmd", "echo 1"),
			vals.MakeMap("id", "5", "cmd", "echo 3"),
			vals.MakeMap("id", "3", "cmd", "echo 2"),
		))

	evals(f.Evaler, `@cmds = (edit:command-history &dedup &newest-first &cmd-only)`)
	testGlobal(t, f.Evaler,
		"cmds",
		vals.MakeList(
			"echo 1",
			"echo 3",
			"echo 2",
		))
}

func TestInsertLastWord(t *testing.T) {
	f := setup(storeOp(func(s store.Store) {
		s.AddCmd("echo foo bar")
	}))
	defer f.Cleanup()

	evals(f.Evaler, "edit:insert-last-word")
	wantBuf := tk.CodeBuffer{Content: "bar", Dot: 3}
	if buf := f.Editor.app.CodeArea().CopyState().Buffer; buf != wantBuf {
		t.Errorf("buf = %v, want %v", buf, wantBuf)
	}
}

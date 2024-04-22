package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/tk"
)

func TestInsertAtDot(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "ab", Dot: 1})
	evals(f.Evaler, `edit:insert-at-dot XYZ`)

	testCodeBuffer(t, f.Editor, tk.CodeBuffer{Content: "aXYZb", Dot: 4})
}

func TestReplaceInput(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "ab", Dot: 1})
	evals(f.Evaler, `edit:replace-input XYZ`)

	testCodeBuffer(t, f.Editor, tk.CodeBuffer{Content: "XYZ", Dot: 3})
}

func TestDot(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "code", Dot: 4})
	evals(f.Evaler, `set edit:-dot = 0`)

	testCodeBuffer(t, f.Editor, tk.CodeBuffer{Content: "code", Dot: 0})
}

func TestDotOutOfBoundary(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "", Dot: 0})
	evals(f.Evaler, "var err = ?(set edit:-dot = 10)[reason]")
	testGlobal(t, f.Evaler, "err", errDotOutOfBoundary)
}

func TestCurrentCommand(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `set edit:current-command = code`)

	testCodeBuffer(t, f.Editor, tk.CodeBuffer{Content: "code", Dot: 4})
}

func testCodeBuffer(t *testing.T, ed *Editor, wantBuf tk.CodeBuffer) {
	t.Helper()
	if buf := codeArea(ed.app).CopyState().Buffer; buf != wantBuf {
		t.Errorf("content = %v, want %v", buf, wantBuf)
	}
}

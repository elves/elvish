package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/cli/el/codearea"
)

func TestInsertAtDot(t *testing.T) {
	ed, _, ev, cleanup := setup()
	defer cleanup()

	cliutil.SetCodeBuffer(ed.app, codearea.Buffer{Content: "ab", Dot: 1})
	evalf(ev, `edit:insert-at-dot XYZ`)

	testCodeBuffer(t, ed, codearea.Buffer{Content: "aXYZb", Dot: 4})
}

func TestReplaceInput(t *testing.T) {
	ed, _, ev, cleanup := setup()
	defer cleanup()

	cliutil.SetCodeBuffer(ed.app, codearea.Buffer{Content: "ab", Dot: 1})
	evalf(ev, `edit:replace-input XYZ`)

	testCodeBuffer(t, ed, codearea.Buffer{Content: "XYZ", Dot: 3})
}

func TestDot(t *testing.T) {
	ed, _, ev, cleanup := setup()
	defer cleanup()

	cliutil.SetCodeBuffer(ed.app, codearea.Buffer{Content: "code", Dot: 4})
	evalf(ev, `edit:-dot = 0`)

	testCodeBuffer(t, ed, codearea.Buffer{Content: "code", Dot: 0})
}

func TestCurrentCommand(t *testing.T) {
	ed, _, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:current-command = code`)

	testCodeBuffer(t, ed, codearea.Buffer{Content: "code", Dot: 4})
}

func testCodeBuffer(t *testing.T, ed *Editor, wantBuf codearea.Buffer) {
	t.Helper()
	if buf := cliutil.GetCodeBuffer(ed.app); buf != wantBuf {
		t.Errorf("content = %v, want %v", buf, wantBuf)
	}
}

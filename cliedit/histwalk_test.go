package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/styled"
)

func TestHistWalk(t *testing.T) {
	st, cleanupStore := store.MustGetTempStore()
	defer cleanupStore()
	st.AddCmd("echo a")

	ed, ttyCtrl, _, cleanup := setupWithStore(st)
	defer cleanup()
	_, _, stop := start(ed)
	defer stop()

	ttyCtrl.Inject(term.K(ui.Up))
	wantBufWalk := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a", styles,
			"   GGGG--",
		)).Newline().SetDotToCursor().
		WriteStyled(styled.MarkLines(
			" HISTORY #1 ", styles,
			"mmmmmmmmmmmm",
		)).Buffer()
	ttyCtrl.TestBuffer(t, wantBufWalk)

	ttyCtrl.Inject(term.K(ui.Enter))
	wantBufDone := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a", styles,
			"   gggg  ",
		)).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufDone)
}

package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/styled"
)

func TestHistWalk(t *testing.T) {
	_, cleanupFs := eval.InTempHome()
	defer cleanupFs()
	st, cleanup := store.MustGetTempStore()
	defer cleanup()
	st.AddCmd("echo a")

	ed, ttyCtrl, _ := setupWithStore(st)
	_, _, stop := start(ed)
	defer stop()

	ttyCtrl.Inject(term.K(ui.Up))
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a", styles,
			"   GGGG--",
		)).Newline().SetDotToCursor().
		WriteStyled(styled.MarkLines(
			" HISTORY #1 ", styles,
			"mmmmmmmmmmmm",
		)).Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

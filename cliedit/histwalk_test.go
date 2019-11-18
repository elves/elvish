package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

func TestHistWalk(t *testing.T) {
	f := setupWithOpt(setupOpt{
		StoreOp: func(s storedefs.Store) {
			s.AddCmd("echo a")
		}})
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Up))
	wantBufWalk := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a", styles,
			"   GGGG--",
		)).SetDotHere().Newline().
		WriteStyled(styled.MarkLines(
			" HISTORY #1 ", styles,
			"mmmmmmmmmmmm",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufWalk)

	f.TTYCtrl.Inject(term.K(ui.Enter))
	wantBufDone := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a", styles,
			"   gggg  ",
		)).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufDone)
}

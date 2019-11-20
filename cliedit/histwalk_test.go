package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
)

func TestHistWalk_Up_EndOfHistory(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Up))
	wantNotesBuf := bb().WritePlain("end of history").Buffer()
	f.TTYCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestHistWalk_Down_EndOfHistory(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	// Not bound by default, so we need to use evals.
	evals(f.Evaler, `edit:history:down`)
	wantNotesBuf := bb().WritePlain("end of history").Buffer()
	f.TTYCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestHistWalk_Accept(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Enter))
	wantBufDone := bb().
		WriteMarkedLines(
			"~> echo a", styles,
			"   gggg  ",
		).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufDone)
}

func TestHistWalk_Close(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('[', ui.Ctrl))
	wantBufClosed := bb().WritePlain("~> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufClosed)
}

func TestHistWalk_DownOrQuit(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Down))
	wantBufClosed := bb().WritePlain("~> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufClosed)
}

func startHistwalkTest(t *testing.T) *fixture {
	// The part of the test shared by all tests.
	f := setupWithOpt(setupOpt{
		StoreOp: func(s storedefs.Store) {
			s.AddCmd("echo a")
		}})

	f.TTYCtrl.Inject(term.K(ui.Up))
	wantBufWalk := bb().
		WriteMarkedLines(
			"~> echo a", styles,
			"   GGGG--",
		).SetDotHere().Newline().
		WriteMarkedLines(
			" HISTORY #1 ", styles,
			"mmmmmmmmmmmm",
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufWalk)
	return f
}

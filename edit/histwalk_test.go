package edit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/ui"
)

func TestHistWalk_Up_EndOfHistory(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Up))
	f.TestTTYNotes(t, "end of history")
}

func TestHistWalk_Down_EndOfHistory(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	// Not bound by default, so we need to use evals.
	evals(f.Evaler, `edit:history:down`)
	f.TestTTYNotes(t, "end of history")
}

func TestHistWalk_Accept(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Right))
	f.TestTTY(t,
		"~> echo a", Styles,
		"   vvvv  ", term.DotHere,
	)
}

func TestHistWalk_Close(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('[', ui.Ctrl))
	f.TestTTY(t, "~> ", term.DotHere)
}

func TestHistWalk_DownOrQuit(t *testing.T) {
	f := startHistwalkTest(t)
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(ui.Down))
	f.TestTTY(t, "~> ", term.DotHere)
}

func TestHistory_FastForward(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo a")
	}))
	defer f.Cleanup()

	f.Store.AddCmd("echo b")
	evals(f.Evaler, `edit:history:fast-forward`)
	f.TTYCtrl.Inject(term.K(ui.Up))
	f.TestTTY(t,
		"~> echo b", Styles,
		"   VVVV__", term.DotHere, "\n",
		" HISTORY #2 ", Styles,
		"************",
	)
}

func startHistwalkTest(t *testing.T) *fixture {
	// The part of the test shared by all tests.
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo a")
	}))

	f.TTYCtrl.Inject(term.K(ui.Up))
	f.TestTTY(t,
		"~> echo a", Styles,
		"   VVVV__", term.DotHere, "\n",
		" HISTORY #1 ", Styles,
		"************",
	)
	return f
}

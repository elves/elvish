package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/ui"
)

/*
func TestInitListing_Binding(t *testing.T) {
	// Test that the binding variable in the returned namespace indeed refers to
	// the BindingMap returned.
	_, binding, ns := initListing(&fakeApp{})
	if ns["binding"].Get() != *binding {
		t.Errorf("The binding var in the ns is not the same as the BindingMap")
	}
}
*/

// Smoke tests for individual addons.

func TestHistlistAddon(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("ls")
		s.AddCmd("echo")
		s.AddCmd("ls")
	}))
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('R', ui.Ctrl))
	wantBuf := bbAddon(" HISTORY (dedup on) ").
		MarkLines(
			"   1 echo\n",
			"   2 ls                       ", styles,
			"##############################",
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)
	wantBuf = bbAddon(" HISTORY ").
		MarkLines(
			"   0 ls\n",
			"   1 echo\n",
			"   2 ls                       ", styles,
			"##############################",
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	evals(f.Evaler, `edit:histlist:toggle-case-sensitivity`)
	wantBuf = bbAddon(" HISTORY (case-insensitive) ").
		MarkLines(
			"   0 ls\n",
			"   1 echo\n",
			"   2 ls                       ", styles,
			"##############################",
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLastCmdAddon(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo hello world")
	}))
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(',', ui.Alt))
	wantBuf := bbAddon("LASTCMD").
		MarkLines(
			"    echo hello world          \n", styles,
			"##############################",
			"  0 echo\n",
			"  1 hello\n",
			"  2 world",
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func bbAddon(name string) *term.BufferBuilder {
	return term.NewBufferBuilder(30).
		Write("~> ").Newline().
		WriteStyled(layout.ModeLine(name, true)).SetDotHere().Newline()
}

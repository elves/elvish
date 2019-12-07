package cliedit

import (
	"testing"

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
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('R', ui.Ctrl))
	f.TestTTY(t,
		"~> \n",
		" HISTORY (dedup on)  ", Styles,
		"******************** ", term.DotHere, "\n",
		"   1 echo\n",
		"   2 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)
	f.TestTTY(t,
		"~> \n",
		" HISTORY  ", Styles,
		"********* ", term.DotHere, "\n",
		"   0 ls\n",
		"   1 echo\n",
		"   2 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-case-sensitivity`)
	f.TestTTY(t,
		"~> \n",
		" HISTORY (case-insensitive)  ", Styles,
		"**************************** ", term.DotHere, "\n",
		"   0 ls\n",
		"   1 echo\n",
		"   2 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)
}

func TestLastCmdAddon(t *testing.T) {
	f := setup(storeOp(func(s storedefs.Store) {
		s.AddCmd("echo hello world")
	}))
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(',', ui.Alt))
	f.TestTTY(t,
		"~> \n",
		"LASTCMD ", Styles,
		"******* ", term.DotHere, "\n",
		"    echo hello world                              \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"  0 echo\n",
		"  1 hello\n",
		"  2 world",
	)
}

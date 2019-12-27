package edit

import (
	"testing"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/ui"
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
	f := setup(storeOp(func(s store.Store) {
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
		"   2 echo\n",
		"   3 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)
	f.TestTTY(t,
		"~> \n",
		" HISTORY  ", Styles,
		"********* ", term.DotHere, "\n",
		"   1 ls\n",
		"   2 echo\n",
		"   3 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-case-sensitivity`)
	f.TestTTY(t,
		"~> \n",
		" HISTORY (case-insensitive)  ", Styles,
		"**************************** ", term.DotHere, "\n",
		"   1 ls\n",
		"   2 echo\n",
		"   3 ls                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)
}

func TestLastCmdAddon(t *testing.T) {
	f := setup(storeOp(func(s store.Store) {
		s.AddCmd("echo hello world")
	}))
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(',', ui.Alt))
	f.TestTTY(t,
		"~> \n",
		" LASTCMD  ", Styles,
		"********* ", term.DotHere, "\n",
		"    echo hello world                              \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"  0 echo\n",
		"  1 hello\n",
		"  2 world",
	)
}

func TestCustomListing_PassingList(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`items = [[&to-filter=1 &to-accept=echo &to-show=echo]
		          [&to-filter=2  &to-accept=put &to-show=(styled put green)]]`,
		`edit:listing:start-custom $items &accept=$edit:insert-at-dot~ &caption=A`)
	f.TestTTY(t,
		"~> \n",
		"A ", Styles,
		"* ", term.DotHere, "\n",
		"echo                                              \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"put                                               ", Styles,
		"vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv",
	)
	// Filter - "put" will be selected.
	f.TTYCtrl.Inject(term.K('2'))
	// Accept.
	f.TTYCtrl.Inject(term.K('\n'))
	f.TestTTY(t,
		"~> put", Styles,
		"   vvv", term.DotHere,
	)
}

func TestCustomListing_PassingValueCallback(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`f = [q]{ put [&to-accept='q '$q &to-show=(styled 'q '$q blue)] }`,
		`edit:listing:start-custom $f &caption=A`)
	// Query.
	f.TTYCtrl.Inject(term.K('x'))
	f.TestTTY(t,
		"~> \n",
		"A x", Styles,
		"*  ", term.DotHere, "\n",
		"q x                                               ", Styles,
		"##################################################",
	)
	// No-op accept.
	f.TTYCtrl.Inject(term.K('\n'))
	f.TestTTY(t, "~> ", term.DotHere)
}

func TestCustomListing_PassingBytesCallback(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`f = [q]{ echo 'q '$q }`,
		`edit:listing:start-custom $f &accept=$edit:insert-at-dot~ &caption=A`)
	// Query.
	f.TTYCtrl.Inject(term.K('x'))
	f.TestTTY(t,
		"~> \n",
		"A x", Styles,
		"*  ", term.DotHere, "\n",
		"q x                                               ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)
}

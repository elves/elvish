package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

// Smoke tests for individual addons.

func TestHistlistAddon(t *testing.T) {
	f := setup(t, storeOp(func(s storedefs.Store) {
		s.AddCmd("ls")
		s.AddCmd("echo")
		s.AddCmd("ls")
		s.AddCmd("LS")
	}))

	f.TTYCtrl.Inject(term.K('R', ui.Ctrl))
	f.TestTTY(t,
		"~> \n",
		" HISTORY (dedup on)  ", Styles,
		"******************** ", term.DotHere, "\n",
		"   2 echo\n",
		"   3 ls\n",
		"   4 LS                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)
	f.TestTTY(t,
		"~> \n",
		" HISTORY  ", Styles,
		"********* ", term.DotHere, "\n",
		"   1 ls\n",
		"   2 echo\n",
		"   3 ls\n",
		"   4 LS                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)

	// Filtering is case-insensitive when filter is all lower case.
	f.TTYCtrl.Inject(term.K('l'))
	f.TestTTY(t,
		"~> \n",
		" HISTORY (dedup on)  l", Styles,
		"********************  ", term.DotHere, "\n",
		"   3 ls\n",
		"   4 LS                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)

	// Filtering is case-sensitive when filter is not all lower case.
	f.TTYCtrl.Inject(term.K(ui.Backspace), term.K('L'))
	f.TestTTY(t,
		"~> \n",
		" HISTORY (dedup on)  L", Styles,
		"********************  ", term.DotHere, "\n",
		"   4 LS                                           ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)
}

func TestLastCmdAddon(t *testing.T) {
	f := setup(t, storeOp(func(s storedefs.Store) {
		s.AddCmd("echo hello world")
	}))

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
	f := setup(t)

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
	f := setup(t)

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
	f := setup(t)

	evals(f.Evaler,
		`f = [q]{ echo '# '$q }`,
		`edit:listing:start-custom $f &accept=$edit:insert-at-dot~ &caption=A `+
			`&binding=(edit:binding-table [&Ctrl-X=$edit:listing:accept~])`)
	// Test that the query function is used to generate candidates. Also test
	// the caption.
	f.TTYCtrl.Inject(term.K('x'))
	f.TestTTY(t,
		"~> \n",
		"A x", Styles,
		"*  ", term.DotHere, "\n",
		"# x                                               ", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
	)
	// Test both the binding and the accept callback.
	f.TTYCtrl.Inject(term.K('X', ui.Ctrl))
	f.TestTTY(t,
		"~> # x", Styles,
		"   ccc", term.DotHere)
}

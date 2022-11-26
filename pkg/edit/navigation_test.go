package edit

import (
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

func TestNavigation(t *testing.T) {
	f := setupNav(t)

	feedInput(f.TTYCtrl, "put")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put", Styles,
		"vvv", term.DotHere, "\n",
		" NAVIGATING            Ctrl-H hidden Ctrl-F filter\n", Styles,
		"************           ++++++        ++++++       ",
		" d      a                 \n", Styles,
		"###### ++++++++++++++++++ ",
		"        e                ", Styles,
		"       //////////////////",
	)

	// Test $edit:selected-file.
	evals(f.Evaler, `var file = $edit:selected-file`)
	wantFile := "a"
	if file, _ := f.Evaler.Global().Index("file"); file != wantFile {
		t.Errorf("Got $edit:selected-file %q, want %q", file, wantFile)
	}

	// Test Alt-Enter: inserts filename without quitting.
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put a", Styles,
		"vvv ", term.DotHere, "\n",
		" NAVIGATING            Ctrl-H hidden Ctrl-F filter\n", Styles,
		"************           ++++++        ++++++       ",
		" d      a                 \n", Styles,
		"###### ++++++++++++++++++ ",
		"        e                ", Styles,
		"       //////////////////",
	)

	// Test Enter: inserts filename and quits.
	f.TTYCtrl.Inject(term.K(ui.Enter))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put a a", Styles,
		"vvv    ", term.DotHere,
	)
}

func TestNavigation_WidthRatio(t *testing.T) {
	f := setupNav(t)

	evals(f.Evaler, `set @edit:navigation:width-ratio = 1 1 1`)
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ", term.DotHere, "\n",
		" NAVIGATING            Ctrl-H hidden Ctrl-F filter\n", Styles,
		"************           ++++++        ++++++       ",
		" d                a               \n", Styles,
		"################ ++++++++++++++++ ",
		"                  e              ", Styles,
		"                 ////////////////",
	)
}

// Test corner case: Inserting a selection when the CLI cursor is not at the
// start of the edit buffer, but the preceding char is a space, does not
// insert another space.
func TestNavigation_EnterDoesNotAddSpaceAfterSpace(t *testing.T) {
	f := setupNav(t)

	feedInput(f.TTYCtrl, "put ")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl)) // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Down))      // select "e"
	f.TTYCtrl.Inject(term.K(ui.Enter))     // insert the "e" file name
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put e", Styles,
		"vvv", term.DotHere,
	)
}

// Test corner case: Inserting a selection when the CLI cursor is at the start
// of the edit buffer omits the space char prefix.
func TestNavigation_EnterDoesNotAddSpaceAtStartOfBuffer(t *testing.T) {
	f := setupNav(t)

	f.TTYCtrl.Inject(term.K('N', ui.Ctrl)) // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Enter))     // insert the "a" file name
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"a", Styles,
		"!", term.DotHere,
	)
}

// Test corner case: Inserting a selection when the CLI cursor is at the start
// of a line buffer omits the space char prefix.
func TestNavigation_EnterDoesNotAddSpaceAtStartOfLine(t *testing.T) {
	f := setupNav(t)

	feedInput(f.TTYCtrl, "put [\n")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl)) // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Enter))     // insert the "a" file name
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put [", Styles,
		"vvv b", "\n",
		"     a", term.DotHere,
	)
}

// Test corner case: Inserting the "selection" in an empty directory inserts
// nothing. Regression test for https://b.elv.sh/1169.
func TestNavigation_EnterDoesNothingInEmptyDir(t *testing.T) {
	f := setupNav(t)

	feedInput(f.TTYCtrl, "pu")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))     // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Down))          // select empty directory "e"
	f.TTYCtrl.Inject(term.K(ui.Right))         // move into "e" directory
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt)) // insert nothing since the dir is empty
	f.TTYCtrl.Inject(term.K('t'))              // user presses 'a'
	f.TestTTY(t,
		filepath.Join("~", "d", "e"), "> ",
		"put", Styles,
		"vvv", term.DotHere, "\n",
		" NAVIGATING            Ctrl-H hidden Ctrl-F filter\n", Styles,
		"************           ++++++        ++++++       ",
		" a                        \n", Styles,
		"                          ",
		" e    ", Styles,
		"######",
	)
}

func TestNavigation_UsesEvalerChdir(t *testing.T) {
	f := setupNav(t)
	afterChdirCalled := false
	f.Evaler.AfterChdir = append(f.Evaler.AfterChdir, func(dir string) {
		afterChdirCalled = true
	})

	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TTYCtrl.Inject(term.K(ui.Down))      // select directory "e"
	f.TTYCtrl.Inject(term.K(ui.Right))     // mode into "e"
	f.TTYCtrl.Inject(term.K('[', ui.Ctrl)) // quit navigation mode

	f.TestTTY(t,
		filepath.Join("~", "d", "e"), "> ", term.DotHere)

	if !afterChdirCalled {
		t.Errorf("afterChdir not called")
	}
}

var testDir = testutil.Dir{
	"d": testutil.Dir{
		"a": "",
		"e": testutil.Dir{},
	},
}

func setupNav(c testutil.Cleanuper) *fixture {
	f := setup(c)
	lscolors.SetTestLsColors(c)
	testutil.ApplyDir(testDir)
	must.Chdir("d")
	return f
}

package edit

import (
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/cli/lscolors"
	"github.com/elves/elvish/pkg/testutil"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

func TestNavigation(t *testing.T) {
	f, cleanup := setupNav()
	defer cleanup()

	feedInput(f.TTYCtrl, "put")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put", Styles,
		"vvv", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d      a                 \n", Styles,
		"###### ++++++++++++++++++ ",
		"        e                ", Styles,
		"       //////////////////",
	)

	// Test $edit:selected-file.
	evals(f.Evaler, `file = $edit:selected-file`)
	wantFile := "a"
	if file := f.Evaler.Global["file"].Get().(string); file != wantFile {
		t.Errorf("Got $edit:selected-file %q, want %q", file, wantFile)
	}

	// Test Alt-Enter: inserts filename without quitting.
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"put a", Styles,
		"vvv ", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
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
	f, cleanup := setupNav()
	defer cleanup()

	evals(f.Evaler, `@edit:navigation:width-ratio = 1 1 1`)
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d                a               \n", Styles,
		"################ ++++++++++++++++ ",
		"                  e              ", Styles,
		"                 ////////////////",
	)
}

// Test corner case: Inserting a selection when the CLI cursor is not at the
// start of the edit buffer, but the preceding char is a space, does not
// insert another space.
func TestNavigationCornerCaseRedundantSpace(t *testing.T) {
	f, cleanup := setupNav()
	defer cleanup()

	f.TTYCtrl.Inject(term.K('x'))              // user presses 'x'
	f.TTYCtrl.Inject(term.K(' '))              // user presses Space
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))     // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Down))          // select "e"
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt)) // insert the "e" file name
	f.TTYCtrl.Inject(term.K('z'))              // user presses 'z'
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"x ez", Styles,
		"!", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d      a                 \n", Styles,
		"######                    ",
		"        e                ", Styles,
		"       ##################",
	)
}

// Test corner case: Inserting a selection when the CLI cursor is at the start
// of the edit buffer omits the space char prefix.
func TestNavigationCornerCaseStartOfBuffer(t *testing.T) {
	f, cleanup := setupNav()
	defer cleanup()

	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))     // begin navigation mode
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt)) // insert the "a" file name
	f.TestTTY(t,
		filepath.Join("~", "d"), "> ",
		"a", Styles,
		"!", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d      a                 \n", Styles,
		"###### ++++++++++++++++++ ",
		"        e                ", Styles,
		"       //////////////////",
	)
}

// Test corner case: Inserting the "selection" in an empty directory inserts
// nothing rather than the literal `''`  that used to be inserted.
func TestNavigationCornerCaseEmptyDir(t *testing.T) {
	f, cleanup := setupNav()
	defer cleanup()

	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))     // begin navigation mode
	f.TTYCtrl.Inject(term.K('z'))              // user presses 'z'
	f.TTYCtrl.Inject(term.K(ui.Down))          // select empty directory "e"
	f.TTYCtrl.Inject(term.K(ui.Right))         // move into "e" directory
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt)) // insert nothing since the dir is empty
	f.TTYCtrl.Inject(term.K('a'))              // user presses 'a'
	f.TTYCtrl.Inject(term.K(' '))              // user presses Space
	f.TestTTY(t,
		filepath.Join("~", "d", "e"), "> ",
		"za ", Styles,
		"!! ", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a                        \n", Styles,
		"                          ",
		" e    ", Styles,
		"######",
	)
}

var testDir = testutil.Dir{
	"d": testutil.Dir{
		"a": "",
		"e": testutil.Dir{},
	},
}

func setupNav() (*fixture, func()) {
	f := setup()
	restoreLsColors := lscolors.WithTestLsColors()

	testutil.ApplyDir(testDir)
	testutil.MustChdir("d")

	return f, func() {
		restoreLsColors()
		f.Cleanup()
	}
}

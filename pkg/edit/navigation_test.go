package edit

import (
	"os"
	"testing"

	"github.com/elves/elvish/pkg/cli/lscolors"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
	"github.com/elves/elvish/pkg/util"
)

func TestNavigation(t *testing.T) {
	f, cleanup := setupNav()
	defer cleanup()

	// Test navigation addon UI.
	feedInput(f.TTYCtrl, "put")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TestTTY(t,
		"~"+string(os.PathSeparator)+"d> ",
		"put", Styles,
		"vvv", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d      a                 ", Styles,
		"###### ++++++++++++++++++ ",
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
		"~"+string(os.PathSeparator)+"d> ",
		"put a", Styles,
		"vvv ", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d      a                 ", Styles,
		"###### ++++++++++++++++++ ",
	)

	// Test Enter: inserts filename and quits.
	f.TTYCtrl.Inject(term.K(ui.Enter))
	f.TestTTY(t,
		"~"+string(os.PathSeparator)+"d> ",
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
		"~"+string(os.PathSeparator)+"d> ", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d                a               ", Styles,
		"################ ++++++++++++++++ ",
	)
}

func setupNav() (*fixture, func()) {
	f := setup()
	restoreLsColors := lscolors.WithTestLsColors()

	util.ApplyDir(util.Dir{"d": util.Dir{"a": ""}})
	err := os.Chdir("d")
	if err != nil {
		panic(err)
	}

	return f, func() {
		restoreLsColors()
		f.Cleanup()
	}
}

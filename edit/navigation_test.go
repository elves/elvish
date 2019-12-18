package edit

import (
	"os"
	"testing"

	"github.com/elves/elvish/cli/lscolors"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
	"github.com/elves/elvish/util"
)

func TestNavigation(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	restoreLsColors := lscolors.WithTestLsColors()
	defer restoreLsColors()

	util.ApplyDir(util.Dir{"d": util.Dir{"a": ""}})
	err := os.Chdir("d")
	if err != nil {
		panic(err)
	}

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

package cliedit

import (
	"os"
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

func TestNavigation(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	restoreEnv := util.WithTempEnv("LS_COLORS", "di=34")
	defer restoreEnv()

	util.ApplyDir(util.Dir{"d": util.Dir{"a": ""}})
	err := os.Chdir("d")
	if err != nil {
		panic(err)
	}

	styles := map[rune]string{
		'#': "blue inverse",
		'-': "inverse",
	}
	makeBuf := func(moreCode string, markedLines ...interface{}) *ui.Buffer {
		b := bb().
			WritePlain("~"+string(os.PathSeparator)+"d> ").
			WriteString("put", "green").
			WritePlain(moreCode).SetDotHere()
		if len(markedLines) > 0 {
			b.Newline().
				WriteStyled(layout.ModeLine(" NAVIGATING ", true)).
				Newline().WriteMarkedLines(markedLines...)
		}
		return b.Buffer()
	}

	// Test navigation addon UI.
	feedInput(f.TTYCtrl, "put")
	f.TTYCtrl.Inject(term.K('N', ui.Ctrl))
	f.TTYCtrl.TestBuffer(t, makeBuf("",
		" d       a                    ", styles,
		"####### --------------------- ",
	))

	// Test $edit:selected-file.
	evals(f.Evaler, `file = $edit:selected-file`)
	wantFile := "a"
	if file := f.Evaler.Global["file"].Get().(string); file != wantFile {
		t.Errorf("Got $edit:selected-file %q, want %q", file, wantFile)
	}

	// Test Alt-Enter: inserts filename without quitting.
	f.TTYCtrl.Inject(term.K(ui.Enter, ui.Alt))
	f.TTYCtrl.TestBuffer(t, makeBuf(" a",
		" d       a                    ", styles,
		"####### --------------------- ",
	))

	// Test Enter: inserts filename and quits.
	f.TTYCtrl.Inject(term.K(ui.Enter))
	f.TTYCtrl.TestBuffer(t, makeBuf(" a a"))
}

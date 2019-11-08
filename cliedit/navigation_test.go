package cliedit

import (
	"os"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestNavigation(t *testing.T) {
	_, ttyCtrl, ev, cleanup := setup()
	defer cleanup()
	util.ApplyDir(util.Dir{"d": util.Dir{"a": ""}})
	err := os.Chdir("d")
	if err != nil {
		panic(err)
	}

	ttyCtrl.Inject(term.K('N', ui.Ctrl))
	styles := map[rune]string{
		'#': "blue inverse",
		'-': "inverse",
	}
	wantBuf := bb().
		WritePlain("~" + string(os.PathSeparator) + "d> ").
		Newline().SetDotToCursor().
		WriteStyled(styled.MarkLines(
			" d       a                    ", styles,
			"####### --------------------- ",
		)).
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	evalf(ev, `file = $edit:selected-file`)
	wantFile := "a"
	if file := ev.Global["file"].Get().(string); file != wantFile {
		t.Errorf("Got $edit:selected-file %q, want %q", file, wantFile)
	}
}

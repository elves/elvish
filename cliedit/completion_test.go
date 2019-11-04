package cliedit

import (
	"testing"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletion(t *testing.T) {
	_, ttyCtrl, _, cleanup := setupStarted()
	defer cleanup()
	util.ApplyDir(util.Dir{"a": "", "b": ""})

	feedInput(ttyCtrl, "echo \t")
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a ", styles,
			"   gggg --",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"a  b", styles,
			"#   ",
		)).
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

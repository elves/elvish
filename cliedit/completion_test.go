package cliedit

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletion(t *testing.T) {
	_, cleanupDir := eval.InTempHome()
	util.ApplyDir(util.Dir{"a": "", "b": ""})
	defer cleanupDir()
	_, ttyCtrl, _, _, cleanup := setupStarted()
	defer cleanup()

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

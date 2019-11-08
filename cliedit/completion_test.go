package cliedit

import (
	"testing"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletion(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"a": "", "b": ""})

	feedInput(f.TTYCtrl, "echo \t")
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
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

package cliedit

import (
	"testing"

	"github.com/elves/elvish/cliedit/complete"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletionAddon(t *testing.T) {
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

func TestCompleteFilename(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"d": util.Dir{"a": "", "b": ""}})

	evals(f.Evaler, `@cands = (edit:complete-filename ls ./d/a)`)
	wantCands := vals.MakeList(
		complete.ComplexItem{Stem: "./d/a", CodeSuffix: " "},
		complete.ComplexItem{Stem: "./d/b", CodeSuffix: " "})
	if cands := getGlobal(f.Evaler, "cands"); !vals.Equal(cands, wantCands) {
		t.Errorf("got cands %s, want %s",
			vals.Repr(cands, vals.NoPretty), vals.Repr(wantCands, vals.NoPretty))
	}
}

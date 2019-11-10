package cliedit

import (
	"testing"

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
	testGlobal(t, f.Evaler,
		"cands",
		vals.MakeList(
			complexItem{Stem: "./d/a", CodeSuffix: " "},
			complexItem{Stem: "./d/b", CodeSuffix: " "}))
}

func TestComplexCandidate(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`cand  = (edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix='x')`,
		// Identical to $cand.
		`cand2 = (edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix='x')`,
		// Different from $cand.
		`cand3 = (edit:complex-candidate a/b/c)`,
		`kind  = (kind-of $cand)`,
		`@keys = (keys $cand)`,
		`repr  = (repr $cand)`,
		`eq2   = (eq $cand $cand2)`,
		`eq2h  = [&$cand=$true][$cand2]`,
		`eq3   = (eq $cand $cand3)`,
		`stem code-suffix display-suffix = $cand[stem code-suffix display-suffix]`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"kind": "map",
		"keys": vals.MakeList("stem", "code-suffix", "display-suffix"),
		"repr": "(edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix=x)",
		"eq2":  true,
		"eq2h": true,
		"eq3":  false,

		"stem":           "a/b/c",
		"code-suffix":    " ",
		"display-suffix": "x",
	})
}

func TestMatchers(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`@prefix = (edit:match-prefix ab [ab abc cab acb ba [ab] [a b] [b a]])`,
		`@substr = (edit:match-substr ab [ab abc cab acb ba [ab] [a b] [b a]])`,
		`@subseq = (edit:match-subseq ab [ab abc cab acb ba [ab] [a b] [b a]])`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"prefix": vals.MakeList(true, true, false, false, false, false, false, false),
		"substr": vals.MakeList(true, true, true, false, false, true, false, false),
		"subseq": vals.MakeList(true, true, true, true, false, true, true, false),
	})
}

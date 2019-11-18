package cliedit

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

func TestCompletionAddon(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"a": "", "b": ""})

	feedInput(f.TTYCtrl, "echo \t")
	wantBuf := bb().
		WriteMarkedLines(
			"~> echo a ", styles,
			"   gggg --",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ").
		SetDotHere().
		Newline().
		WriteMarkedLines(
			"a  b", styles,
			"#   ",
		).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestCompletionAddon_CompletesLongestCommonPrefix(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"foo1": "", "foo2": "", "foo": "", "fox": ""})

	feedInput(f.TTYCtrl, "echo \t")
	wantBuf := bb().
		WriteMarkedLines(
			"~> echo fo", styles,
			"   gggg").
		SetDotHere().
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	feedInput(f.TTYCtrl, "\t")
	wantBuf = bb().
		WriteMarkedLines(
			"~> echo foo ", styles,
			"   gggg ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ").
		SetDotHere().
		Newline().
		WriteMarkedLines(
			"foo  foo1  foo2  fox", styles,
			"###                 ",
		).
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

func TestCompletionArgCompleter_ArgsAndValueOutput(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`foo-args = []`,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   foo-args = $args
		   put 1val
		   edit:complex-candidate 2val &display-suffix=_
		 }`)

	feedInput(f.TTYCtrl, "foo foo1 foo2 \t")
	wantBuf := bb().
		WriteMarkedLines(
			"~> foo foo1 foo2 1val", styles,
			"   ggg           ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ").
		SetDotHere().
		Newline().
		WriteMarkedLines(
			"1val  2val_", styles,
			"####       ",
		).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
	testGlobal(t, f.Evaler,
		"foo-args", vals.MakeList("foo", "foo1", "foo2", ""))
}

func TestCompletionArgCompleter_BytesOutput(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   echo 1val
		   echo 2val
		 }`)

	feedInput(f.TTYCtrl, "foo foo1 foo2 \t")
	wantBuf := bb().
		WriteMarkedLines(
			"~> foo foo1 foo2 1val", styles,
			"   ggg           ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ").
		SetDotHere().
		Newline().
		WriteMarkedLines(
			"1val  2val", styles,
			"####      ",
		).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestCompleteSudo(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   echo val1
		   echo val2
		 }`,
		`@cands = (edit:complete-sudo sudo foo '')`)
	testGlobal(t, f.Evaler, "cands", vals.MakeList("val1", "val2"))
}

func TestCompletionMatcher(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"foo": "", "oof": ""})

	evals(f.Evaler, `edit:completion:matcher[''] = $edit:match-substr~`)
	feedInput(f.TTYCtrl, "echo f\t")
	wantBuf := bb().
		WriteMarkedLines(
			"~> echo foo ", styles,
			"   gggg ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ").
		SetDotHere().
		Newline().
		WriteMarkedLines(
			"foo  oof", styles,
			"###     ",
		).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestBuiltinMatchers(t *testing.T) {
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

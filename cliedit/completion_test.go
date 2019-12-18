package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

func TestCompletionAddon(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"a": "", "b": ""})

	feedInput(f.TTYCtrl, "echo \t")
	f.TestTTY(t,
		"~> echo a \n", Styles,
		"   vvvv __",
		" COMPLETING argument  ", Styles,
		"********************* ", term.DotHere, "\n",
		"a  b", Styles,
		"+   ",
	)
}

func TestCompletionAddon_CompletesLongestCommonPrefix(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"foo1": "", "foo2": "", "foo": "", "fox": ""})

	feedInput(f.TTYCtrl, "echo \t")
	f.TestTTY(t,
		"~> echo fo", Styles,
		"   vvvv", term.DotHere,
	)

	feedInput(f.TTYCtrl, "\t")
	f.TestTTY(t,
		"~> echo foo \n", Styles,
		"   vvvv ____",
		" COMPLETING argument  ", Styles,
		"********************* ", term.DotHere, "\n",
		"foo  foo1  foo2  fox", Styles,
		"+++                 ",
	)
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
	f.TestTTY(t,
		"~> foo foo1 foo2 1val\n", Styles,
		"   vvv           ____",
		" COMPLETING argument  ", Styles,
		"********************* ", term.DotHere, "\n",
		"1val  2val_", Styles,
		"++++       ",
	)
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
	f.TestTTY(t,
		"~> foo foo1 foo2 1val\n", Styles,
		"   vvv           ____",
		" COMPLETING argument  ", Styles,
		"********************* ", term.DotHere, "\n",
		"1val  2val", Styles,
		"++++      ",
	)
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
	f.TestTTY(t,
		"~> echo foo \n", Styles,
		"   vvvv ____",
		" COMPLETING argument  ", Styles,
		"********************* ", term.DotHere, "\n",
		"foo  oof", Styles,
		"+++     ",
	)
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

func TestBuiltinMatchers_Options(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	// The two options work identically on all the builtin matchers, so we only
	// test for match-prefix for simplicity.
	evals(f.Evaler,
		`@a = (edit:match-prefix &ignore-case ab [abc aBc AbC])`,
		`@b = (edit:match-prefix &ignore-case aB [abc aBc AbC])`,
		`@c = (edit:match-prefix &smart-case  ab [abc aBc Abc])`,
		`@d = (edit:match-prefix &smart-case  aB [abc aBc AbC])`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"a": vals.MakeList(true, true, true),
		"b": vals.MakeList(true, true, true),
		"c": vals.MakeList(true, true, true),
		"d": vals.MakeList(false, true, false),
	})
}

package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/tt"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestClosureAsValue(t *testing.T) {
	Test(t,
		// Basic operations as a value.
		That("kind-of { }").Puts("fn"),
		That("eq { } { }").Puts(false),
		That("var x = { }; put [&$x= foo][$x]").Puts("foo"),

		// Argument arity mismatch.
		That("var f = {|x| }", "$f a b").Throws(
			errs.ArityMismatch{What: "arguments",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"$f a b"),
		That("var f = {|x y| }", "$f a").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 2, ValidHigh: 2, Actual: 1},
			"$f a"),
		That("var f = {|x y @rest| }", "$f a").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 2, ValidHigh: -1, Actual: 1},
			"$f a"),

		// Unsupported option.
		That("var f = {|&valid1=1 &valid2=2| }; $f &bad1=1 &bad2=2").Throws(
			eval.UnsupportedOptionsError{[]string{"bad1", "bad2"}},
			"$f &bad1=1 &bad2=2"),

		That("all {|a b| }[arg-names]").Puts("a", "b"),
		That("put {|@r| }[rest-arg]").Puts("0"),
		That("all {|&opt=def| }[opt-names]").Puts("opt"),
		That("all {|&opt=def| }[opt-defaults]").Puts("def"),
		That("put { body }[body]").Puts("body "),
		That("put {|x @y| body }[def]").Puts("{|x @y| body }"),
		That("put { body }[src][code]").
			Puts("put { body }[src][code]"),

		// Regression test for https://b.elv.sh/1126
		That("fn f { body }; put $f~[body]").Puts("body "),
	)
}

func TestUnsupportedOptionsError(t *testing.T) {
	tt.Test(t, error.Error,
		tt.Args(eval.UnsupportedOptionsError{[]string{"sole-opt"}}).
			Rets("unsupported option: sole-opt"),
		tt.Args(eval.UnsupportedOptionsError{[]string{"opt-foo", "opt-bar"}}).
			Rets("unsupported options: opt-foo, opt-bar"),
	)
}

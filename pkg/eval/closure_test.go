package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/errs"
)

func TestClosure(t *testing.T) {
	Test(t,
		That("kind-of { }").Puts("fn"),
		That("eq { } { }").Puts(false),
		That("x = { }; put [&$x= foo][$x]").Puts("foo"),

		That("f = [x]{ }", "$f a b").Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"$f a b"),
		That("f = [x y]{ }", "$f a").Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 2, ValidHigh: 2, Actual: 1},
			"$f a"),
		That("f = [x y @rest]{ }", "$f a").Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 2, ValidHigh: -1, Actual: 1},
			"$f a"),

		That("[]{ } &k=v").ThrowsAny(),

		That("all [a b]{ }[arg-names]").Puts("a", "b"),
		That("put [@r]{ }[rest-arg]").Puts("0"),
		That("all [&opt=def]{ }[opt-names]").Puts("opt"),
		That("all [&opt=def]{ }[opt-defaults]").Puts("def"),
		That("put { body }[body]").Puts(" body "),
		That("put [x @y]{ body }[def]").Puts("[x @y]{ body }"),
		That("put { body }[src][code]").
			Puts("put { body }[src][code]"),

		// Regression test for https://b.elv.sh/1126
		That("fn f { body }; put $f~[body]").Puts(" body "),
	)
}

package eval

import "testing"

func TestClosure(t *testing.T) {
	runTests(t, []Test{
		NewTest("kind-of { }").WantOutStrings("fn"),
		NewTest("eq { } { }").WantOutBools(false),
		NewTest("x = { }; put [&$x= foo][$x]").WantOutStrings("foo"),
		NewTest("[x]{ } a b").WantAnyErr(),
		NewTest("[x y]{ } a").WantAnyErr(),
		NewTest("[x y @rest]{ } a").WantAnyErr(),
		NewTest("[]{ } &k=v").WantAnyErr(),

		NewTest("explode [a b]{ }[arg-names]").WantOutStrings("a", "b"),
		NewTest("put [@r]{ }[rest-arg]").WantOutStrings("r"),
		NewTest("explode [&opt=def]{ }[opt-names]").WantOutStrings("opt"),
		NewTest("explode [&opt=def]{ }[opt-defaults]").WantOutStrings("def"),
		NewTest("put { body }[src][code]").WantOutStrings(
			"put { body }[src][code]"),
		NewTest("put { body }[body]").WantOutStrings(" body "),
	})
}

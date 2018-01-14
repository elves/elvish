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
	})
}

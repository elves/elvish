package eval

import "testing"

func TestClosure(t *testing.T) {
	runTests(t, []Test{
		NewTest("kind-of { }").WantOutStrings("fn"),
		NewTest("eq { } { }").WantOutBools(false),
		NewTest("x = { }; put [&$x= foo][$x]").WantOutStrings("foo"),
	})
}

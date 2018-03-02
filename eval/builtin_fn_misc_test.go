package eval

import "testing"

func TestBuiltinFnMisc(t *testing.T) {
	runTests(t, []Test{
		NewTest("resolve for").WantOutStrings("special"),
		NewTest("resolve put").WantOutStrings("$put~"),
		NewTest("fn f { }; resolve f").WantOutStrings("$f~"),
		NewTest("use lorem; resolve lorem:put-name").WantOutStrings(
			"$lorem:put-name~"),
		NewTest("resolve cat").WantOutStrings("(external cat)"),
	})
}

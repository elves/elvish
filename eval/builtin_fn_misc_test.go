package eval

import "testing"

func TestBuiltinFnMisc(t *testing.T) {
	runTests(t, []Test{
		NewTest("resolve for").WantBytesOutString("special"),
		NewTest("resolve put").WantBytesOutString("$put~"),
		NewTest("fn f { }; resolve f").WantBytesOutString("$f~"),
		NewTest("use lorem; resolve lorem:put-name").WantBytesOutString(
			"$lorem:put-name~"),
		NewTest("resolve cat").WantBytesOutString("(external cat)"),
	})
}

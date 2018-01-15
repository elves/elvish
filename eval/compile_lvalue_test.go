package eval

import "testing"

func TestAssignment(t *testing.T) {
	runTests(t, []Test{
		NewTest("x = a; put $x").WantOutStrings("a"),
		NewTest("x = [a]; x[0] = b; put $x[0]").WantOutStrings("b"),
		NewTest("x = a; { x = b }; put $x").WantOutStrings("b"),
		NewTest("x = [a]; { x[0] = b }; put $x[0]").WantOutStrings("b"),
	})
}

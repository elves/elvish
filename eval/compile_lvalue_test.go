package eval

import "testing"

func TestAssignment(t *testing.T) {
	Test(t,
		That("x = a; put $x").Puts("a"),
		That("x = [a]; x[0] = b; put $x[0]").Puts("b"),
		That("x = a; { x = b }; put $x").Puts("b"),
		That("x = [a]; { x[0] = b }; put $x[0]").Puts("b"),
		// temporary variable
		That("x=ok put $x").Puts("ok"),
		That("x=ok put $x; put $x").Puts("ok").Errors(), // variable does not exist
		// closure
		That("f=[m]{ put $m } { $f ok }").Puts("ok"),
	)
}

package eval

import "testing"

func TestAssignment(t *testing.T) {
	runTests(t, []Test{
		That("x = a; put $x").Puts("a"),
		That("x = [a]; x[0] = b; put $x[0]").Puts("b"),
		That("x = a; { x = b }; put $x").Puts("b"),
		That("x = [a]; { x[0] = b }; put $x[0]").Puts("b"),
	})
}

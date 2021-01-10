package eval_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	. "github.com/elves/elvish/pkg/eval/evaltest"
)

func TestAssignment(t *testing.T) {
	Test(t,
		That("x = a; put $x").Puts("a"),
		That("x = [a]; x[0] = b; put $x[0]").Puts("b"),
		That("x = a; { x = b }; put $x").Puts("b"),
		That("x = [a]; { x[0] = b }; put $x[0]").Puts("b"),

		// Trying to add a new name in a namespace throws an exception.
		// Regression test for #1214.
		That("ns: = (ns [&]); ns:a = b").Throws(NoSuchVariable("ns:a"), "ns:a = b"),
	)
}

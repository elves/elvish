package vals

import (
	"bytes"
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

func TestToString(t *testing.T) {
	Test(t, Fn("ToString", ToString), Table{
		// string
		Args("a").Rets("a"),

		// float64
		Args(42.0).Rets("42"),
		Args(0.1).Rets("0.1"),
		// Exponents >= 14 are printed in scientific notation.
		Args(1e13).Rets("10000000000000"),
		Args(1e14).Rets("1e+14"),
		// Exponents <= -5 are printed in scientific notation.
		Args(1e-4).Rets("0.0001"),
		Args(1e-5).Rets("1e-05"),

		// Stringer
		Args(bytes.NewBufferString("buffer")).Rets("buffer"),
		// None of the above: delegate to Repr
		Args(true).Rets("$true"),
	})
}

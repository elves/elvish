package vals

import (
	"bytes"
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestToString(t *testing.T) {
	tt.Test(t, ToString,
		// string
		Args("a").Rets("a"),

		Args(1).Rets("1"),

		// float64
		Args(0.1).Rets("0.1"),
		Args(42.0).Rets("42.0"),
		// Whole numbers with more than 14 digits and trailing 0 are printed in
		// scientific notation.
		Args(1e13).Rets("10000000000000.0"),
		Args(1e14).Rets("1e+14"),
		Args(1e14+1).Rets("100000000000001.0"),
		// Numbers smaller than 0.0001 are printed in scientific notation.
		Args(0.0001).Rets("0.0001"),
		Args(0.00001).Rets("1e-05"),
		Args(0.00009).Rets("9e-05"),

		// Stringer
		Args(bytes.NewBufferString("buffer")).Rets("buffer"),
		// None of the above: delegate to Repr
		Args(true).Rets("$true"),
	)
}

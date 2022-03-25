package vals

import (
	"bytes"
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestToString(t *testing.T) {
	tt.Test(t, tt.Fn("ToString", ToString), tt.Table{
		// string
		tt.Args("a").Rets("a"),

		tt.Args(1).Rets("1"),

		// float64
		tt.Args(0.1).Rets("0.1"),
		tt.Args(42.0).Rets("42.0"),
		// Whole numbers with more than 14 digits and trailing 0 are printed in
		// scientific notation.
		tt.Args(1e13).Rets("10000000000000.0"),
		tt.Args(1e14).Rets("1e+14"),
		tt.Args(1e14 + 1).Rets("100000000000001.0"),
		// Numbers smaller than 0.0001 are printed in scientific notation.
		tt.Args(0.0001).Rets("0.0001"),
		tt.Args(0.00001).Rets("1e-05"),
		tt.Args(0.00009).Rets("9e-05"),

		// Stringer
		tt.Args(bytes.NewBufferString("buffer")).Rets("buffer"),
		// None of the above: delegate to Repr
		tt.Args(true).Rets("$true"),
	})
}

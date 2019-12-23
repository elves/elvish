package vals

import (
	"bytes"
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestToString(t *testing.T) {
	tt.Test(t, tt.Fn("ToString", ToString), tt.Table{
		// string
		tt.Args("a").Rets("a"),
		// float64
		tt.Args(42.0).Rets("42"),
		tt.Args(0.1).Rets("0.1"),
		// Stringer
		tt.Args(bytes.NewBufferString("buffer")).Rets("buffer"),
		// None of the above: delegate to Repr
		tt.Args(true).Rets("$true"),
	})
}

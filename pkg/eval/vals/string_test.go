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
		// Stringer
		Args(bytes.NewBufferString("buffer")).Rets("buffer"),
		// None of the above: delegate to Repr
		Args(true).Rets("$true"),
	})
}

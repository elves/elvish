package vals

import (
	"os"
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

type xtype int

func TestKind(t *testing.T) {
	Test(t, Fn("Kind", Kind), Table{
		Args(nil).Rets("nil"),
		Args(true).Rets("bool"),
		Args("").Rets("string"),
		Args(os.Stdin).Rets("file"),
		Args(EmptyList).Rets("list"),
		Args(EmptyMap).Rets("map"),

		Args(xtype(0)).Rets("!!vals.xtype"),

		Args(os.Stdin).Rets("file"),
	})
}

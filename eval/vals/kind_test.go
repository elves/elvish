package vals

import (
	"os"
	"testing"

	"github.com/elves/elvish/tt"
)

type xtype int

func TestKind(t *testing.T) {
	tt.Test(t, tt.Fn("Kind", Kind), tt.Table{
		Args(nil).Rets("nil"),
		Args(true).Rets("bool"),
		Args("").Rets("string"),
		Args(os.Stdin).Rets("file"),
		Args(EmptyList).Rets("list"),
		Args(EmptyMap).Rets("map"),
		Args(testStructMap{}).Rets("structmap"),

		Args(xtype(0)).Rets("!!vals.xtype"),

		Args(NewStruct(NewStructDescriptor(), nil)).Rets("map"),
		Args(os.Stdin).Rets("file"),
		Args(NewPipe(os.Stdin, os.Stdout)).Rets("pipe"),
	})
}

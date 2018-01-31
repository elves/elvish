package types

import (
	"os"
	"testing"

	"github.com/elves/elvish/tt"
)

type xtype int

func TestKind(t *testing.T) {
	tt.Test(t, tt.Fn("Kind", Kind), tt.Table{
		Args(true).Rets("bool"),
		Args("").Rets("string"),
		Args(EmptyList).Rets("list"),
		Args(EmptyMap).Rets("map"),
		Args(xtype(0)).Rets("!!types.xtype"),

		Args(NewStruct(NewStructDescriptor(), nil)).Rets("map"),
		Args(NewFile(os.Stdin)).Rets("file"),
		Args(NewPipe(os.Stdin, os.Stdout)).Rets("pipe"),
	})
}

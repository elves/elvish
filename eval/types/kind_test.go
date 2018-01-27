package types

import (
	"os"
	"testing"

	"github.com/elves/elvish/tt"
	"github.com/xiaq/persistent/vector"
)

func TestKind(t *testing.T) {
	tt.Test(t, tt.Fn("Kind", Kind), tt.Table{
		Args(true).Rets("bool"),
		Args("").Rets("string"),
		Args(NewList(vector.Empty)).Rets("list"),
		Args(NewMap(EmptyMapInner)).Rets("map"),
		Args(NewStruct(NewStructDescriptor(), nil)).Rets("map"),
		Args(NewFile(os.Stdin)).Rets("file"),
		Args(NewPipe(os.Stdin, os.Stdout)).Rets("pipe"),
	})
}

package types

import (
	"os"
	"testing"

	"github.com/elves/elvish/tt"
	"github.com/xiaq/persistent/vector"
)

var Args = tt.Args

func kind(k Kinder) string {
	return k.Kind()
}

func TestKind(t *testing.T) {
	tt.Test(t, tt.Fn("kind", kind), tt.Table{
		Args(Bool(true)).Rets("bool"),
		Args(String("")).Rets("string"),
		Args(NewList(vector.Empty)).Rets("list"),
		Args(NewMap(EmptyMapInner)).Rets("map"),
		Args(NewStruct(NewStructDescriptor(), nil)).Rets("map"),
		Args(NewFile(os.Stdin)).Rets("file"),
		Args(NewPipe(os.Stdin, os.Stdout)).Rets("pipe"),
	})
}

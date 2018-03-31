package vals

import (
	"testing"

	"github.com/elves/elvish/tt"
)

type customEqualer struct{ ret bool }

func (c customEqualer) Equal(interface{}) bool { return c.ret }

type customStruct struct{ a, b string }

var equalTests = tt.Table{
	Args(customEqualer{true}, 2).Rets(true),
	Args(customEqualer{false}, 2).Rets(false),
	Args(true, true).Rets(true),
	Args(true, false).Rets(false),
	Args(JsonNil{}, JsonNil{}).Rets(true),
	Args(true, JsonNil{}).Rets(false),
	Args(false, JsonNil{}).Rets(false),
	Args("lorem", "lorem").Rets(true),
	Args("lorem", "ipsum").Rets(false),
	Args(MakeList("a", "b"), MakeList("a", "b")).Rets(true),
	Args(MakeList("a", "b"), MakeList("a")).Rets(false),
	Args(MakeList("a", "b"), MakeList("a", "c")).Rets(false),
	Args(MakeMapFromKV("k", "v"), MakeMapFromKV("k", "v")).Rets(true),
	Args(MakeMapFromKV("k", "v"), MakeMapFromKV("k2", "v")).Rets(false),
	Args(MakeMapFromKV("k", "v", "k2", "v2"), MakeMapFromKV("k", "v")).Rets(false),
	Args(&customStruct{"a", "b"}, &customStruct{"a", "b"}).Rets(true),
}

func TestEqual(t *testing.T) {
	tt.Test(t, tt.Fn("Equal", Equal), equalTests)
}

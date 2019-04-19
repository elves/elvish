package vals

import (
	"os"
	"testing"

	"github.com/elves/elvish/tt"
)

type customEqualer struct{ ret bool }

func (c customEqualer) Equal(interface{}) bool { return c.ret }

type customStruct struct{ a, b string }

var equalTests = tt.Table{
	Args(nil, nil).Rets(true),
	Args(nil, "").Rets(false),

	Args(true, true).Rets(true),
	Args(true, false).Rets(false),

	Args(1.0, 1.0).Rets(true),

	Args("lorem", "lorem").Rets(true),
	Args("lorem", "ipsum").Rets(false),

	Args(os.Stdin, os.Stdin).Rets(true),
	Args(os.Stdin, os.Stderr).Rets(false),

	Args(MakeList("a", "b"), MakeList("a", "b")).Rets(true),
	Args(MakeList("a", "b"), MakeList("a")).Rets(false),
	Args(MakeList("a", "b"), MakeList("a", "c")).Rets(false),

	Args(MakeMap("k", "v"), MakeMap("k", "v")).Rets(true),
	Args(MakeMap("k", "v"), MakeMap("k2", "v")).Rets(false),
	Args(MakeMap("k", "v", "k2", "v2"), MakeMap("k", "v")).Rets(false),

	Args(testStructMap{}, testStructMap{}).Rets(true),
	Args(testStructMap{"a", 1.0}, testStructMap{"a", 1.0}).Rets(true),
	Args(testStructMap{"a", 1.0}, testStructMap{"a", 2.0}).Rets(false),
	Args(testStructMap{"a", 1.0}, "a").Rets(false),

	Args(customEqualer{true}, 2).Rets(true),
	Args(customEqualer{false}, 2).Rets(false),

	Args(&customStruct{"a", "b"}, &customStruct{"a", "b"}).Rets(true),
}

func TestEqual(t *testing.T) {
	tt.Test(t, tt.Fn("Equal", Equal), equalTests)
}

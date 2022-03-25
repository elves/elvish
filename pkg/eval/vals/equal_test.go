package vals

import (
	"math/big"
	"os"
	"testing"

	"src.elv.sh/pkg/tt"
)

type customEqualer struct{ ret bool }

func (c customEqualer) Equal(any) bool { return c.ret }

type customStruct struct{ a, b string }

func TestEqual(t *testing.T) {
	tt.Test(t, tt.Fn("Equal", Equal), tt.Table{
		tt.Args(nil, nil).Rets(true),
		tt.Args(nil, "").Rets(false),

		tt.Args(true, true).Rets(true),
		tt.Args(true, false).Rets(false),

		tt.Args(1.0, 1.0).Rets(true),
		tt.Args(1.0, 1.1).Rets(false),
		tt.Args("1.0", 1.0).Rets(false),
		tt.Args(1, 1.0).Rets(false),
		tt.Args(1, 1).Rets(true),
		tt.Args(bigInt(z), bigInt(z)).Rets(true),
		tt.Args(bigInt(z), 1).Rets(false),
		tt.Args(bigInt(z), bigInt(z1)).Rets(false),
		tt.Args(big.NewRat(1, 2), big.NewRat(1, 2)).Rets(true),
		tt.Args(big.NewRat(1, 2), 0.5).Rets(false),

		tt.Args("lorem", "lorem").Rets(true),
		tt.Args("lorem", "ipsum").Rets(false),

		tt.Args(os.Stdin, os.Stdin).Rets(true),
		tt.Args(os.Stdin, os.Stderr).Rets(false),
		tt.Args(os.Stdin, "").Rets(false),
		tt.Args(os.Stdin, 0).Rets(false),

		tt.Args(MakeList("a", "b"), MakeList("a", "b")).Rets(true),
		tt.Args(MakeList("a", "b"), MakeList("a")).Rets(false),
		tt.Args(MakeList("a", "b"), MakeList("a", "c")).Rets(false),
		tt.Args(MakeList("a", "b"), "").Rets(false),
		tt.Args(MakeList("a", "b"), 1.0).Rets(false),

		tt.Args(MakeMap("k", "v"), MakeMap("k", "v")).Rets(true),
		tt.Args(MakeMap("k", "v"), MakeMap("k2", "v")).Rets(false),
		tt.Args(MakeMap("k", "v", "k2", "v2"), MakeMap("k", "v")).Rets(false),
		tt.Args(MakeMap("k", "v"), "").Rets(false),
		tt.Args(MakeMap("k", "v"), 1.0).Rets(false),

		tt.Args(customEqualer{true}, 2).Rets(true),
		tt.Args(customEqualer{false}, 2).Rets(false),

		tt.Args(&customStruct{"a", "b"}, &customStruct{"a", "b"}).Rets(true),
	})
}

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
	tt.Test(t, Equal,
		Args(nil, nil).Rets(true),
		Args(nil, "").Rets(false),

		Args(true, true).Rets(true),
		Args(true, false).Rets(false),

		Args(1.0, 1.0).Rets(true),
		Args(1.0, 1.1).Rets(false),
		Args("1.0", 1.0).Rets(false),
		Args(1, 1.0).Rets(false),
		Args(1, 1).Rets(true),
		Args(bigInt(z), bigInt(z)).Rets(true),
		Args(bigInt(z), 1).Rets(false),
		Args(bigInt(z), bigInt(z1)).Rets(false),
		Args(big.NewRat(1, 2), big.NewRat(1, 2)).Rets(true),
		Args(big.NewRat(1, 2), 0.5).Rets(false),

		Args("lorem", "lorem").Rets(true),
		Args("lorem", "ipsum").Rets(false),

		Args(os.Stdin, os.Stdin).Rets(true),
		Args(os.Stdin, os.Stderr).Rets(false),
		Args(os.Stdin, "").Rets(false),
		Args(os.Stdin, 0).Rets(false),

		Args(MakeList("a", "b"), MakeList("a", "b")).Rets(true),
		Args(MakeList("a", "b"), MakeList("a")).Rets(false),
		Args(MakeList("a", "b"), MakeList("a", "c")).Rets(false),
		Args(MakeList("a", "b"), "").Rets(false),
		Args(MakeList("a", "b"), 1.0).Rets(false),

		Args(MakeMap("k", "v"), MakeMap("k", "v")).Rets(true),
		Args(MakeMap("k", "v"), MakeMap("k2", "v")).Rets(false),
		Args(MakeMap("k", "v", "k2", "v2"), MakeMap("k", "v")).Rets(false),
		Args(MakeMap("k", "v"), "").Rets(false),
		Args(MakeMap("k", "v"), 1.0).Rets(false),

		Args(customEqualer{true}, 2).Rets(true),
		Args(customEqualer{false}, 2).Rets(false),

		Args(&customStruct{"a", "b"}, &customStruct{"a", "b"}).Rets(true),
	)
}

package vals

import (
	"math/big"
	"os"
	"testing"

	"src.elv.sh/pkg/tt"
)

type xtype int

func TestKind(t *testing.T) {
	tt.Test(t, tt.Fn("Kind", Kind), tt.Table{
		tt.Args(nil).Rets("nil"),
		tt.Args(true).Rets("bool"),
		tt.Args("").Rets("string"),
		tt.Args(1).Rets("number"),
		tt.Args(bigInt(z)).Rets("number"),
		tt.Args(big.NewRat(1, 2)).Rets("number"),
		tt.Args(1.0).Rets("number"),
		tt.Args(os.Stdin).Rets("file"),
		tt.Args(EmptyList).Rets("list"),
		tt.Args(EmptyMap).Rets("map"),
		tt.Args(xtype(0)).Rets("!!vals.xtype"),
		tt.Args(os.Stdin).Rets("file"),
	})
}

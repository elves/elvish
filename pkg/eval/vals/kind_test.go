package vals

import (
	"math/big"
	"os"
	"testing"

	"src.elv.sh/pkg/tt"
)

type xtype int

func TestKind(t *testing.T) {
	tt.Test(t, Kind,
		Args(nil).Rets("nil"),
		Args(true).Rets("bool"),
		Args("").Rets("string"),
		Args(1).Rets("number"),
		Args(bigInt(z)).Rets("number"),
		Args(big.NewRat(1, 2)).Rets("number"),
		Args(1.0).Rets("number"),
		Args(os.Stdin).Rets("file"),
		Args(EmptyList).Rets("list"),
		Args(EmptyMap).Rets("map"),
		Args(xtype(0)).Rets("!!vals.xtype"),
		Args(os.Stdin).Rets("file"),
	)
}

package vals

import (
	"math/big"
	"os"
	"testing"

	. "src.elv.sh/pkg/tt"
)

type structMap struct{}

func (structMap) IsStructMap() {}

type xtype int

func TestKind(t *testing.T) {
	// This does not exercise the Kinder case. We instead rely on unit tests for the types which
	// implement that interface to exercise that case.
	Test(t, Fn("Kind", Kind), Table{
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
		Args(structMap{}).Rets("structmap"),
		Args(xtype(0)).Rets("!!vals.xtype"),
	})
}

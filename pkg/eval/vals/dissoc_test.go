package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type dissocer struct{}

func (dissocer) Dissoc(any) any { return "custom ret" }

func TestDissoc(t *testing.T) {
	tt.Test(t, tt.Fn("Dissoc", Dissoc), tt.Table{
		Args(MakeMap("k1", "v1", "k2", "v2"), "k1").Rets(eq(MakeMap("k2", "v2"))),
		Args(dissocer{}, "x").Rets("custom ret"),
		Args("", "x").Rets(nil),
	})
}

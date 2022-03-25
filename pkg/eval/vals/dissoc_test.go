package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type dissocer struct{}

func (dissocer) Dissoc(any) any { return "custom ret" }

func TestDissoc(t *testing.T) {
	tt.Test(t, tt.Fn("Dissoc", Dissoc), tt.Table{
		tt.Args(MakeMap("k1", "v1", "k2", "v2"), "k1").Rets(eq(MakeMap("k2", "v2"))),
		tt.Args(dissocer{}, "x").Rets("custom ret"),
		tt.Args("", "x").Rets(nil),
	})
}

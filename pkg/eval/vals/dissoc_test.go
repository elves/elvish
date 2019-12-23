package vals

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

type dissocer struct{}

func (dissocer) Dissoc(interface{}) interface{} { return "custom ret" }

var dissocTests = tt.Table{
	Args(MakeMap("k1", "v1", "k2", "v2"), "k1").Rets(
		eq(MakeMap("k2", "v2"))),
	Args(dissocer{}, "x").Rets("custom ret"),
	Args("", "x").Rets(nil),
}

func TestDissoc(t *testing.T) {
	tt.Test(t, tt.Fn("Dissoc", Dissoc), dissocTests)
}

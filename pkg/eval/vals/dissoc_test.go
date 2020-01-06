package vals

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

type dissocer struct{}

func (dissocer) Dissoc(interface{}) interface{} { return "custom ret" }

func TestDissoc(t *testing.T) {
	Test(t, Fn("Dissoc", Dissoc), Table{
		Args(MakeMap("k1", "v1", "k2", "v2"), "k1").Rets(Eq(MakeMap("k2", "v2"))),
		Args(dissocer{}, "x").Rets("custom ret"),
		Args("", "x").Rets(nil),
	})
}

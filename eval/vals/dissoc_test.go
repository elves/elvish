package vals

import (
	"testing"

	"github.com/elves/elvish/tt"
)

type customDissocer struct{}

func (customDissocer) Dissoc(interface{}) interface{} { return "custom ret" }

var dissocTests = tt.Table{
	Args(MakeMapFromKV("k1", "v1", "k2", "v2"), "k1").Rets(
		eq(MakeMapFromKV("k2", "v2"))),
	Args(customDissocer{}, "x").Rets("custom ret"),
}

func TestDissoc(t *testing.T) {
	tt.Test(t, tt.Fn("Dissoc", Dissoc), dissocTests)
}

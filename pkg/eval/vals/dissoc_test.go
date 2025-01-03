package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type dissocer struct{}

func (dissocer) Dissoc(any) any { return "custom ret" }

func TestDissoc(t *testing.T) {
	tt.Test(t, Dissoc,
		Args(MakeMap("k1", "v1", "k2", "v2"), "k1").
			Rets(eq(MakeMap("k2", "v2"))),
		Args(fieldMap{"lorem", "ipsum", 23}, "foo-bar").
			Rets(eq(MakeMap("foo", "lorem", "bar", "ipsum"))),
		Args(dissocer{}, "x").Rets("custom ret"),
		Args("", "x").Rets(nil),
	)
}

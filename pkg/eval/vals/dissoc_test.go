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
		Args(testStructMap{"ls", 1.0}, "score-plus-ten").
			Rets(eq(MakeMap("name", "ls", "score", 1.0))),
		Args(dissocer{}, "x").Rets("custom ret"),
		Args("", "x").Rets(nil),
	)
}

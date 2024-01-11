package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestCmp(t *testing.T) {
	// Cmp is tested by tests of the Elvish compare command.
}

func TestCmpTotal_StructMap(t *testing.T) {
	// CmpTotal should pretend that structmaps are maps too. Since maps don't
	// have an internal ordering, comparing a structmap to another structmap or
	// to a map should always return CmpEqual, like comparing two maps.
	//
	// This is not covered by tests of the Elvish compare command because Elvish
	// code are not supposed to know which values are actually structmaps.
	x := testStructMap{}
	y := testStructMap2{}
	z := EmptyMap
	tt.Test(t, CmpTotal,
		tt.Args(x, y).Rets(CmpEqual),
		tt.Args(x, z).Rets(CmpEqual),
	)
}

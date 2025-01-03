package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestCmp(t *testing.T) {
	// Cmp is tested by tests of the Elvish compare command.
}

func TestCmpTotal_FieldMap(t *testing.T) {
	// CmpTotal should pretend that field maps are maps too. Since maps don't
	// have an internal ordering, comparing a field map to another field map or
	// to a map should always return CmpEqual, like comparing two maps.
	//
	// This is not covered by tests of the Elvish compare command because Elvish
	// code are not supposed to know which values are actually field maps.
	x := fieldMap{}
	y := fieldMap{}
	z := EmptyMap
	tt.Test(t, CmpTotal,
		tt.Args(x, y).Rets(CmpEqual),
		tt.Args(x, z).Rets(CmpEqual),
	)
}

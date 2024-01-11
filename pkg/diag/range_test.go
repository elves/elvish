package diag

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

type aRanger struct {
	Ranging
}

func TestEmbeddingRangingImplementsRanger(t *testing.T) {
	r := Ranging{1, 10}
	s := Ranger(aRanger{Ranging{1, 10}})
	if s.Range() != r {
		t.Errorf("s.Range() = %v, want %v", s.Range(), r)
	}
}

func TestPointRanging(t *testing.T) {
	tt.Test(t, PointRanging,
		Args(1).Rets(Ranging{1, 1}),
	)
}

func TestMixedRanging(t *testing.T) {
	tt.Test(t, MixedRanging,
		Args(Ranging{1, 2}, Ranging{0, 4}).Rets(Ranging{1, 4}),
		Args(Ranging{0, 4}, Ranging{1, 2}).Rets(Ranging{0, 2}),
	)
}

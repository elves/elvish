package diag

import "testing"

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

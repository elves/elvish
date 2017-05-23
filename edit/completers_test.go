package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/eval"
)

func TestComplIndexInner(t *testing.T) {
	m := eval.NewMap(map[eval.Value]eval.Value{
		eval.String("foo"):   eval.String("bar"),
		eval.String("lorem"): eval.String("ipsum"),
	})
	wantCandidates := []rawCandidate{
		plainCandidate("foo"), plainCandidate("lorem"),
	}
	candidates := complIndexInner(m)
	if !reflect.DeepEqual(candidates, wantCandidates) {
		t.Errorf("complIndexInner(%v) = %v, want %v",
			m, candidates, wantCandidates)
	}
}

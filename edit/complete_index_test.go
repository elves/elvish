package edit

import (
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
)

var testIndexee = eval.String("a")

func TestFindIndexComplContext(t *testing.T) {
	testComplContextFinder(t, "findIndexComplContext", findIndexComplContext, []complContextFinderTest{
		{"a[", &indexComplContext{
			complContextCommon{"", quotingForEmptySeed, 2, 2}, testIndexee}},
		{"a[x", &indexComplContext{
			complContextCommon{"x", parse.Bareword, 2, 3}, testIndexee}},
		{"a[x ", &indexComplContext{
			complContextCommon{"", quotingForEmptySeed, 4, 4}, testIndexee}},
		// Not supported when indexee cannot be evaluated statically
		{"(x)[", nil},
		// Multi-layer indexing not supported yet
		{"a[x][", nil},
	})
}

func TestComplIndexInner(t *testing.T) {
	m := eval.ConvertToMap(map[types.Value]types.Value{
		eval.String("foo"):   eval.String("bar"),
		eval.String("lorem"): eval.String("ipsum"),
	})
	var (
		candidates     rawCandidates
		wantCandidates = rawCandidates{
			plainCandidate("foo"), plainCandidate("lorem"),
		}
	)

	gets := make(chan rawCandidate)
	go func() {
		defer close(gets)
		complIndexInner(m, gets)
	}()
	for v := range gets {
		candidates = append(candidates, v)
	}
	sort.Sort(candidates)
	if !reflect.DeepEqual(candidates, wantCandidates) {
		t.Errorf("complIndexInner(%v) = %v, want %v",
			m, candidates, wantCandidates)
	}
}

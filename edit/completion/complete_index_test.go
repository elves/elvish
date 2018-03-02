package completion

import (
	"testing"

	"github.com/elves/elvish/parse"
)

var testIndexee = "a"

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

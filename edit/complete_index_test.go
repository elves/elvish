package edit

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

var testIndexee = eval.String("a")

func TestFindIndexComplContext(t *testing.T) {
	testComplContextFinder(t, "findIndexComplContext", findIndexComplContext, []complContextFinderTest{
		{"a[", &indexComplContext{"", quotingForEmptySeed, testIndexee, 2, 2}},
		{"a[x", &indexComplContext{"x", parse.Bareword, testIndexee, 2, 3}},
		{"a[x ", &indexComplContext{"", quotingForEmptySeed, testIndexee, 4, 4}},
		// Not supported when indexee cannot be evaluated statically
		{"(x)[", nil},
		// Multi-layer indexing not supported yet
		{"a[x][", nil},
	})
}

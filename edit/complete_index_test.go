package edit

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

var testIndexee = eval.String("a")

func TestFindIndexCompleter(t *testing.T) {
	testCompleterFinder(t, "findIndexCompleter", findIndexCompleter, []completerFinderTest{
		{"a[", &indexCompleter{"", quotingForEmptySeed, testIndexee, 2, 2}},
		{"a[x", &indexCompleter{"x", parse.Bareword, testIndexee, 2, 3}},
		{"a[x ", &indexCompleter{"", quotingForEmptySeed, testIndexee, 4, 4}},
		// Not supported when indexee cannot be evaluated statically
		{"(x)[", nil},
		// Multi-layer indexing not supported yet
		{"a[x][", nil},
	})
}

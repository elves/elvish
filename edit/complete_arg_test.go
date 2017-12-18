package edit

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindArgCompleter(t *testing.T) {
	testCompleterFinder(t, "findArgCompleter", findArgCompleter, []completerFinderTest{
		{"a ", &argCompleter{"", quotingForEmptySeed, []string{"a", ""}, 2, 2}},
		{"a b", &argCompleter{"b", parse.Bareword, []string{"a", "b"}, 2, 3}},
		// No space after command; won't complete arg
		{"a", nil},
	})
}

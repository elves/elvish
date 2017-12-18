package edit

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindArgComplContext(t *testing.T) {
	testComplContextFinder(t, "findArgComplContext", findArgComplContext, []complContextFinderTest{
		{"a ", &argComplContext{"", quotingForEmptySeed, []string{"a", ""}, 2, 2}},
		{"a b", &argComplContext{"b", parse.Bareword, []string{"a", "b"}, 2, 3}},
		// No space after command; won't complete arg
		{"a", nil},
	})
}

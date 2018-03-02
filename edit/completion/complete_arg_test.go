package completion

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindArgComplContext(t *testing.T) {
	testComplContextFinder(t, "findArgComplContext", findArgComplContext, []complContextFinderTest{
		{"a ", &argComplContext{
			complContextCommon{"", quotingForEmptySeed, 2, 2}, []string{"a", ""}}},
		{"a b", &argComplContext{
			complContextCommon{"b", parse.Bareword, 2, 3}, []string{"a", "b"}}},
		// No space after command; won't complete arg
		{"a", nil},
	})
}

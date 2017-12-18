package edit

import "testing"

func TestFindRedirComplContext(t *testing.T) {
	testComplContextFinder(t, "findRedirComplContext", findRedirComplContext, []complContextFinderTest{
		{"a >", &redirComplContext{"", quotingForEmptySeed, 3, 3}},
		{"a >b", &redirComplContext{"b", quotingForEmptySeed, 3, 4}},
	})
}

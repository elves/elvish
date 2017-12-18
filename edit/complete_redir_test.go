package edit

import "testing"

func TestFindRedirComplContext(t *testing.T) {
	testComplContextFinder(t, "findRedirComplContext", findRedirComplContext, []complContextFinderTest{
		{"a >", &redirComplContext{
			complContextCommon{"", quotingForEmptySeed, 3, 3}}},
		{"a >b", &redirComplContext{
			complContextCommon{"b", quotingForEmptySeed, 3, 4}}},
	})
}

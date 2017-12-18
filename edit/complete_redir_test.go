package edit

import "testing"

func TestFindRedirCompleter(t *testing.T) {
	testCompleterFinder(t, "findRedirCompleter", findRedirCompleter, []completerFinderTest{
		{"a >", &redirCompleter{"", quotingForEmptySeed, 3, 3}},
		{"a >b", &redirCompleter{"b", quotingForEmptySeed, 3, 4}},
	})
}

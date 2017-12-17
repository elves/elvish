package edit

import "testing"

func TestFindVariableCompleter(t *testing.T) {
	testCompleterFinder(t, "findVariableCompleter", findVariableCompleter, []completerFinderTest{
		{"$", &variableCompleter{"", "", "", 1, 1}},
		{"$a", &variableCompleter{"", "", "a", 1, 2}},
		{"$a:", &variableCompleter{"a", "a:", "", 3, 3}},
		{"$a:b", &variableCompleter{"a", "a:", "b", 3, 4}},
		// Wrong contexts
		{"", nil},
		{"echo", nil},
	})
}

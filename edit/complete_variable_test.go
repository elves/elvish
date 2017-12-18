package edit

import "testing"

func TestFindVariableComplContext(t *testing.T) {
	testComplContextFinder(t, "findVariableComplContext", findVariableComplContext, []complContextFinderTest{
		{"$", &variableComplContext{"", "", "", 1, 1}},
		{"$a", &variableComplContext{"", "", "a", 1, 2}},
		{"$a:", &variableComplContext{"a", "a:", "", 3, 3}},
		{"$a:b", &variableComplContext{"a", "a:", "b", 3, 4}},
		// Wrong contexts
		{"", nil},
		{"echo", nil},
	})
}

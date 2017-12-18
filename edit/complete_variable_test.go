package edit

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindVariableComplContext(t *testing.T) {
	testComplContextFinder(t, "findVariableComplContext", findVariableComplContext, []complContextFinderTest{
		{"$", &variableComplContext{
			complContextCommon{"", parse.Bareword, 1, 1}, "", ""}},
		{"$a", &variableComplContext{
			complContextCommon{"a", parse.Bareword, 1, 2}, "", ""}},
		{"$a:", &variableComplContext{
			complContextCommon{"", parse.Bareword, 3, 3}, "a", "a:"}},
		{"$a:b", &variableComplContext{
			complContextCommon{"b", parse.Bareword, 3, 4}, "a", "a:"}},
		// Wrong contexts
		{"", nil},
		{"echo", nil},
	})
}

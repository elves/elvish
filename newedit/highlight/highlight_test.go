package highlight

import (
	"testing"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

type errorsMatcher struct{ matchers []tt.Matcher }

func (m errorsMatcher) Match(v tt.RetValue) bool {
	errs := v.([]error)
	if len(errs) != len(m.matchers) {
		return false
	}
	for i, matcher := range m.matchers {
		if !matcher.Match(errs[i]) {
			return false
		}
	}
	return true
}

func matchErrors(m ...tt.Matcher) errorsMatcher { return errorsMatcher{m} }

type parseErrorMatcher struct{ begin, end int }

func (m parseErrorMatcher) Match(v tt.RetValue) bool {
	err := v.(*parse.Error)
	return m.begin == err.Context.Begin && m.end == err.Context.End
}

type anyMatcher struct{}

func (anyMatcher) Match(tt.RetValue) bool { return true }

func TestHighlight(t *testing.T) {
	var noErrors []error
	any := anyMatcher{}

	tt.Test(t, tt.Fn("Highlight", Highlight), tt.Table{
		Args("ls").Rets(styled.Text{
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
		}, noErrors),
		Args(" ls\n").Rets(styled.Text{
			styled.UnstyledSegment(" "),
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
			styled.UnstyledSegment("\n"),
		}, noErrors),
		// Parse error
		Args("ls ]").Rets(any, matchErrors(parseErrorMatcher{3, 4})),
		// Errors at the end are elided
		Args("ls $").Rets(any, noErrors),
		Args("ls [").Rets(any, noErrors),

		// TODO: Test for multiple parse errors
	})
}

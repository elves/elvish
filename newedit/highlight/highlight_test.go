package highlight

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var noErrors []error

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
	any := anyMatcher{}

	hl := func(code string) (styled.Text, []error) {
		return highlight(code, Dep{}, nopLateCb)
	}

	tt.Test(t, tt.Fn("highlight", hl), tt.Table{
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

type fakeCheckError struct {
	from, to int
}

func (e fakeCheckError) Range() diag.Ranging {
	return diag.Ranging{e.from, e.to}
}

func (fakeCheckError) Error() string {
	return "fake check error"
}

func TestHighlight_Check(t *testing.T) {
	var checkError error
	dep := Dep{
		Check: func(n *parse.Chunk) error {
			return checkError
		},
	}

	checkError = fakeCheckError{0, 2}
	_, errors := highlight("code", dep, nopLateCb)
	if !reflect.DeepEqual(errors, []error{checkError}) {
		t.Errorf("Got errors %v, want %v", errors, []error{checkError})
	}

	// Errors at the end
	checkError = fakeCheckError{4, 4}
	_, errors = highlight("code", dep, nopLateCb)
	if len(errors) != 0 {
		t.Errorf("Got errors %v, want 0 error", errors)
	}
}

func TestHighlight_HasCommand(t *testing.T) {
	hasCommand := func(cmd string) bool { return cmd == "ls" }
	hl := func(code string) (styled.Text, []error) {
		return highlight(code, Dep{HasCommand: hasCommand}, nopLateCb)
	}

	tt.Test(t, tt.Fn("highlight", hl), tt.Table{
		Args("ls").Rets(styled.Text{
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
		}, noErrors),
		Args("echo").Rets(styled.Text{
			&styled.Segment{styled.Style{Foreground: "red"}, "echo"},
		}, noErrors),
	})
}

func nopLateCb(styled.Text) {}

package highlight

import (
	"reflect"
	"testing"
	"time"

	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/tt"
	"github.com/elves/elvish/ui"
)

var any = anyMatcher{}
var noErrors []error

var styles = ui.RuneStylesheet{
	'?':  ui.BgRed,
	'$':  ui.Magenta,
	'\'': ui.Yellow,
	'v':  ui.Green,
}

func TestHighlighter_HighlightRegions(t *testing.T) {
	hl := NewHighlighter(Config{})

	tt.Test(t, tt.Fn("hl.Get", hl.Get), tt.Table{
		Args("ls").Rets(
			ui.MarkLines(
				"ls", styles,
				"vv",
			),
			noErrors),
		Args(" ls\n").Rets(
			ui.MarkLines(
				" ls\n", styles,
				" vv"),
			noErrors),
		Args("ls $x 'y'").Rets(
			ui.MarkLines(
				"ls $x 'y'", styles,
				"vv $$ '''"),
			noErrors),
	})
}

func TestHighlighter_ParseErrors(t *testing.T) {
	hl := NewHighlighter(Config{})
	tt.Test(t, tt.Fn("hl.Get", hl.Get), tt.Table{
		// Parse error is highlighted and returned
		Args("ls ]").Rets(
			ui.MarkLines(
				"ls ]", styles,
				"vv ?"),
			matchErrors(parseErrorMatcher{3, 4})),
		// Errors at the end are ignored
		Args("ls $").Rets(any, noErrors),
		Args("ls [").Rets(any, noErrors),
	})
}

func TestHighlighter_CheckErrors(t *testing.T) {
	var checkError error
	// Make a highlighter whose Check callback returns checkError.
	hl := NewHighlighter(Config{
		Check: func(*parse.Chunk) error { return checkError }})
	getWithCheckError := func(code string, err error) (ui.Text, []error) {
		checkError = err
		return hl.Get(code)
	}

	tt.Test(t, tt.Fn("getWithCheckError", getWithCheckError), tt.Table{
		// Check error is highlighted and returned
		Args("code 1", fakeCheckError{5, 6}).Rets(
			ui.MarkLines(
				"code 1", styles,
				"vvvv ?"),
			[]error{fakeCheckError{5, 6}}),
		// Check errors at the end are ignored
		Args("code 2", fakeCheckError{6, 6}).
			Rets(any, noErrors),
	})
}

type c struct {
	given       string
	wantInitial ui.Text
	wantLate    ui.Text
	mustLate    bool
}

const lateTimeout = 100 * time.Millisecond

func testThat(t *testing.T, hl *Highlighter, c c) {
	initial, _ := hl.Get(c.given)
	if !reflect.DeepEqual(c.wantInitial, initial) {
		t.Errorf("want %v from initial Get, got %v", c.wantInitial, initial)
	}
	if c.wantLate == nil {
		return
	}
	select {
	case late := <-hl.LateUpdates():
		if !reflect.DeepEqual(c.wantLate, late) {
			t.Errorf("want %v from LateUpdates, got %v", c.wantLate, late)
		}
		late, _ = hl.Get(c.given)
		if !reflect.DeepEqual(c.wantLate, late) {
			t.Errorf("want %v from late Get, got %v", c.wantLate, late)
		}
	case <-time.After(lateTimeout):
		t.Errorf("want %v from LateUpdates, but timed out after %v",
			c.wantLate, lateTimeout)
	}
}

func TestHighlighter_HasCommand_LateResult_Async(t *testing.T) {
	// When the HasCommand callback takes longer than maxBlockForLate, late
	// results are delivered asynchronously.
	maxBlockForLate = 1 * time.Millisecond
	hl := NewHighlighter(Config{
		// HasCommand is slow and only recognizes "ls".
		HasCommand: func(cmd string) bool {
			time.Sleep(10 * time.Millisecond)
			return cmd == "ls"
		}})

	testThat(t, hl, c{
		given:       "ls",
		wantInitial: ui.T("ls"),
		wantLate:    ui.T("ls", ui.Green),
	})
	testThat(t, hl, c{
		given:       "echo",
		wantInitial: ui.T("echo"),
		wantLate:    ui.T("echo", ui.Red),
	})
}

func TestHighlighter_HasCommand_LateResult_Sync(t *testing.T) {
	// When the HasCommand callback takes shorter than maxBlockForLate, late
	// results are delivered asynchronously.
	maxBlockForLate = 100 * time.Millisecond
	hl := NewHighlighter(Config{
		// HasCommand is fast and only recognizes "ls".
		HasCommand: func(cmd string) bool {
			time.Sleep(1 * time.Millisecond)
			return cmd == "ls"
		}})

	testThat(t, hl, c{
		given:       "ls",
		wantInitial: ui.T("ls", ui.Green),
	})
	testThat(t, hl, c{
		given:       "echo",
		wantInitial: ui.T("echo", ui.Red),
	})
}

func TestHighlighter_HasCommand_LateResultOutOfOrder(t *testing.T) {
	// When late results are delivered out of order, the ones that do not match
	// the current code are dropped. In this test, hl.Get is called with "l"
	// first and then "ls". The late result for "l" is delivered after that of
	// "ls" and is dropped.

	// Make sure that the HasCommand callback takes longer than maxBlockForLate.
	maxBlockForLate = 1 * time.Millisecond

	hlSecond := make(chan struct{})
	hl := NewHighlighter(Config{
		HasCommand: func(cmd string) bool {
			if cmd == "l" {
				// Make sure that the second highlight has been requested before
				// returning.
				<-hlSecond
				time.Sleep(10 * time.Millisecond)
				return false
			}
			time.Sleep(10 * time.Millisecond)
			close(hlSecond)
			return cmd == "ls"
		}})

	hl.Get("l")

	initial, _ := hl.Get("ls")
	late := <-hl.LateUpdates()

	wantInitial := ui.T("ls")
	wantLate := ui.T("ls", ui.Green)
	if !reflect.DeepEqual(wantInitial, initial) {
		t.Errorf("want %v from initial Get, got %v", wantInitial, initial)
	}
	if !reflect.DeepEqual(wantLate, late) {
		t.Errorf("want %v from late Get, got %v", wantLate, late)
	}

	// Make sure that no more late updates are delivered.
	select {
	case late := <-hl.LateUpdates():
		t.Errorf("want nothing from LateUpdates, got %v", late)
	case <-time.After(50 * time.Millisecond):
		// We have waited for 50 ms and there are no late updates; test passes.
	}
}

// Matchers.

type anyMatcher struct{}

func (anyMatcher) Match(tt.RetValue) bool { return true }

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

// Fake check error, used in tests for check callback.
type fakeCheckError struct{ from, to int }

func (e fakeCheckError) Range() diag.Ranging { return diag.Ranging{e.from, e.to} }
func (fakeCheckError) Error() string         { return "fake check error" }

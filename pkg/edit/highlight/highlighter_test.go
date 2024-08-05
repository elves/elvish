package highlight

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

var any = anyMatcher{}
var noTips []ui.Text

var styles = ui.RuneStylesheet{
	'?':  ui.Stylings(ui.FgBrightWhite, ui.BgRed),
	'$':  ui.FgMagenta,
	'\'': ui.FgYellow,
	'v':  ui.FgGreen,
}

func TestHighlighter_HighlightRegions(t *testing.T) {
	// Force commands to be delivered synchronously.
	testutil.Set(t, &maxBlockForLate, testutil.Scaled(100*time.Millisecond))
	hl := NewHighlighter(Config{
		HasCommand: func(name string) bool { return name == "ls" },
	})

	tt.Test(t, tt.Fn(hl.Get).Named("hl.Get"),
		Args("ls").Rets(
			ui.MarkLines(
				"ls", styles,
				"vv",
			),
			noTips),
		Args(" ls\n").Rets(
			ui.MarkLines(
				" ls\n", styles,
				" vv"),
			noTips),
		Args("ls $x 'y'").Rets(
			ui.MarkLines(
				"ls $x 'y'", styles,
				"vv $$ '''"),
			noTips),
		// Non-bareword commands do not go through command highlighting.
		Args("'ls'").Rets(ui.T("'ls'", ui.FgYellow)),
		Args("a$x").Rets(
			ui.MarkLines(
				"a$x", styles,
				" $$"),
			noTips,
		),
	)
}

func TestHighlighter_ParseErrors(t *testing.T) {
	hl := NewHighlighter(Config{})
	tt.Test(t, tt.Fn(hl.Get).Named("hl.Get"),
		// Parse error is highlighted and returned
		Args("ls ]").Rets(
			ui.MarkLines(
				"ls ]", styles,
				"vv ?"),
			matchTexts("1:4")),
		// Multiple parse errors
		Args("ls $? ]").Rets(
			ui.MarkLines(
				"ls $? ]", styles,
				"vv $? ?"),
			matchTexts("1:5", "1:7")),
		// Errors at the end are ignored
		Args("ls $").Rets(any, noTips),
		Args("ls [").Rets(any, noTips),
	)
}

func TestHighlighter_AutofixesAndCheckErrors(t *testing.T) {
	ev := eval.NewEvaler()
	ev.AddModule("mod1", &eval.Ns{})
	hl := NewHighlighter(Config{
		Check: func(t parse.Tree) (string, []*eval.CompilationError) {
			autofixes, err := ev.CheckTree(t, nil)
			return strings.Join(autofixes, "; "), eval.UnpackCompilationErrors(err)
		},
		AutofixTip: func(s string) ui.Text { return ui.T("autofix: " + s) },
	})

	tt.Test(t, tt.Fn(hl.Get).Named("hl.Get"),
		// Check error is highlighted and returned
		Args("ls $a ").Rets(
			ui.MarkLines(
				"ls $a ", styles,
				"vv ?? "),
			matchTexts("1:4")),
		// Multiple check errors
		Args("ls $a $b ").Rets(
			ui.MarkLines(
				"ls $a $b ", styles,
				"vv ?? ?? "),
			matchTexts("1:4", "1:7")),
		// Check errors at the end are ignored
		Args("set _").Rets(any, noTips),

		// Autofix
		Args("nop $mod1:").Rets(
			ui.MarkLines(
				"nop $mod1:", styles,
				"vvv $$$$$$"),
			matchTexts(
				"autofix: use mod1", // autofix
			)),
	)
}

type c struct {
	given       string
	wantInitial ui.Text
	wantLate    ui.Text
	mustLate    bool
}

var lateTimeout = testutil.Scaled(100 * time.Millisecond)

func testThat(t *testing.T, hl *Highlighter, c c) {
	initial, _ := hl.Get(c.given)
	if !reflect.DeepEqual(c.wantInitial, initial) {
		t.Errorf("want %v from initial Get, got %v", c.wantInitial, initial)
	}
	if c.wantLate == nil {
		return
	}
	select {
	case <-hl.LateUpdates():
		late, _ := hl.Get(c.given)
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
	testutil.Set(t, &maxBlockForLate, testutil.Scaled(time.Millisecond))
	hl := NewHighlighter(Config{
		// HasCommand is slow and only recognizes "ls".
		HasCommand: func(cmd string) bool {
			time.Sleep(testutil.Scaled(10 * time.Millisecond))
			return cmd == "ls"
		}})

	testThat(t, hl, c{
		given:       "ls",
		wantInitial: ui.T("ls"),
		wantLate:    ui.T("ls", ui.FgGreen),
	})
	testThat(t, hl, c{
		given:       "echo",
		wantInitial: ui.T("echo"),
		wantLate:    ui.T("echo", ui.FgRed),
	})
}

func TestHighlighter_HasCommand_LateResult_Sync(t *testing.T) {
	// When the HasCommand callback takes shorter than maxBlockForLate, late
	// results are delivered asynchronously.
	testutil.Set(t, &maxBlockForLate, testutil.Scaled(100*time.Millisecond))
	hl := NewHighlighter(Config{
		// HasCommand is fast and only recognizes "ls".
		HasCommand: func(cmd string) bool {
			time.Sleep(testutil.Scaled(time.Millisecond))
			return cmd == "ls"
		}})

	testThat(t, hl, c{
		given:       "ls",
		wantInitial: ui.T("ls", ui.FgGreen),
	})
	testThat(t, hl, c{
		given:       "echo",
		wantInitial: ui.T("echo", ui.FgRed),
	})
}

func TestHighlighter_HasCommand_LateResultOutOfOrder(t *testing.T) {
	// When late results are delivered out of order, the ones that do not match
	// the current code are dropped. In this test, hl.Get is called with "l"
	// first and then "ls". The late result for "l" is delivered after that of
	// "ls" and is dropped.

	// Make sure that the HasCommand callback takes longer than maxBlockForLate.
	testutil.Set(t, &maxBlockForLate, testutil.Scaled(time.Millisecond))

	hlSecond := make(chan struct{})
	hl := NewHighlighter(Config{
		HasCommand: func(cmd string) bool {
			if cmd == "l" {
				// Make sure that the second highlight has been requested before
				// returning.
				<-hlSecond
				time.Sleep(testutil.Scaled(10 * time.Millisecond))
				return false
			}
			time.Sleep(testutil.Scaled(10 * time.Millisecond))
			close(hlSecond)
			return cmd == "ls"
		}})

	hl.Get("l")

	testThat(t, hl, c{
		given:       "ls",
		wantInitial: ui.T("ls"),
		wantLate:    ui.T("ls", ui.FgGreen),
		mustLate:    true,
	})

	// Make sure that no more late updates are delivered.
	select {
	case late := <-hl.LateUpdates():
		t.Errorf("want nothing from LateUpdates, got %v", late)
	case <-time.After(testutil.Scaled(50 * time.Millisecond)):
		// We have waited for 50 ms and there are no late updates; test passes.
	}
}

// Matchers.

type anyMatcher struct{}

func (anyMatcher) Match(tt.RetValue) bool { return true }

type textsMatcher struct{ substrings []string }

func matchTexts(s ...string) textsMatcher { return textsMatcher{s} }

func (m textsMatcher) Match(v tt.RetValue) bool {
	texts := v.([]ui.Text)
	if len(texts) != len(m.substrings) {
		return false
	}
	for i, text := range texts {
		if !strings.Contains(text.String(), m.substrings[i]) {
			return false
		}
	}
	return true
}

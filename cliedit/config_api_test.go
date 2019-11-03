package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/term"
)

func TestInitAPI_BeforeReadline(t *testing.T) {
	ed, _, ev, _, cleanup := setup()
	defer cleanup()

	evalf(ev, `called = 0`)
	evalf(ev, `edit:before-readline = [ { called = (+ $called 1) } ]`)

	_, _, stop := start(ed)
	stop()

	// TODO(xiaq): Test more precisely when before-readline is called.
	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
}

func TestInitAPI_AfterReadline(t *testing.T) {
	ed, _, ev, _, cleanup := setup()
	defer cleanup()

	evalf(ev, `called = 0`)
	evalf(ev, `called-with = ''`)
	evalf(ev, `edit:after-readline = [
	             [code]{ called = (+ $called 1); called-with = $code } ]`)

	ed.app.CodeArea().MutateState(func(s *codearea.State) {
		s.CodeBuffer.InsertAtDot("test code")
	})
	_, _, stop := start(ed)
	stop()

	// TODO(xiaq): Test more precisely when after-readline is called.
	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
	if calledWith := ev.Global["called-with"].Get(); calledWith != "test code" {
		t.Errorf("called = %q, want %q", calledWith, "test code")
	}
}

func TestInitAPI_Insert_Abbr(t *testing.T) {
	ed, ttyCtrl, ev, _, cleanup := setup()
	defer cleanup()
	codeCh, _, stop := start(ed)
	defer stop()

	evalf(ev, `edit:insert:abbr = [&x=full]`)
	ttyCtrl.Inject(term.K('x'), term.K('\n'))

	if code := <-codeCh; code != "full" {
		t.Errorf("abbreviation expanded to %q, want %q", code, "full")
	}
}

func TestInitAPI_Insert_Binding(t *testing.T) {
	ed, ttyCtrl, ev, _, cleanup := setup()
	defer cleanup()

	evalf(ev, `called = 0`)
	evalf(ev, `edit:insert:binding[x] = { called = (+ $called 1) }`)

	codeCh, _, _ := start(ed)
	ttyCtrl.Inject(term.K('x'), term.K('\n'))
	code := <-codeCh

	if code != "" {
		t.Errorf("code = %q, want %q", code, "")
	}
	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
}

func TestInitAPI_Insert_QuotePaste(t *testing.T) {
	ed, ttyCtrl, ev, _, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:insert:quote-paste = $true`)

	codeCh, _, _ := start(ed)
	ttyCtrl.Inject(
		term.PasteSetting(true),
		term.K('>'),
		term.PasteSetting(false),
		term.K('\n'))
	code := <-codeCh

	wantCode := `'>'`
	if code != wantCode {
		t.Errorf("Got code %q, want %q", code, wantCode)
	}
}

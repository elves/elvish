package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
)

func TestInsert_Abbr(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
	defer cleanup()
	codeCh, _, stop := start(ed)
	defer stop()

	evalf(ev, `edit:abbr = [&x=full]`)
	ttyCtrl.Inject(term.K('x'), term.K('\n'))

	if code := <-codeCh; code != "full" {
		t.Errorf("abbreviation expanded to %q, want %q", code, "full")
	}
}

func TestInsert_Binding(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
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

func TestInsert_QuotePaste(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
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

func TestToggleQuotePaste(t *testing.T) {
	_, _, ev, cleanup := setupUnstarted()
	defer cleanup()

	evalf(ev, `v0 = $edit:insert:quote-paste`)
	evalf(ev, `edit:toggle-quote-paste`)
	evalf(ev, `v1 = $edit:insert:quote-paste`)
	evalf(ev, `edit:toggle-quote-paste`)
	evalf(ev, `v2 = $edit:insert:quote-paste`)

	v0 := ev.Global["v0"].Get().(bool)
	v1 := ev.Global["v1"].Get().(bool)
	v2 := ev.Global["v2"].Get().(bool)
	if v1 == v0 {
		t.Errorf("got v1 = v0")
	}
	if v2 == v1 {
		t.Errorf("got v2 = v1")
	}
}

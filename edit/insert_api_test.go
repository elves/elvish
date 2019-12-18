package edit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
)

func TestInsert_Abbr(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler, `edit:abbr = [&x=full]`)
	f.TTYCtrl.Inject(term.K('x'), term.K('\n'))

	if code := <-f.codeCh; code != "full" {
		t.Errorf("abbreviation expanded to %q, want %q", code, "full")
	}
}

func TestInsert_Binding(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`called = 0`,
		`edit:insert:binding[x] = { called = (+ $called 1) }`)

	f.TTYCtrl.Inject(term.K('x'), term.K('\n'))

	if code := <-f.codeCh; code != "" {
		t.Errorf("code = %q, want %q", code, "")
	}
	if called := f.Evaler.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
}

func TestInsert_QuotePaste(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler, `edit:insert:quote-paste = $true`)

	f.TTYCtrl.Inject(
		term.PasteSetting(true),
		term.K('>'),
		term.PasteSetting(false),
		term.K('\n'))

	wantCode := `'>'`
	if code := <-f.codeCh; code != wantCode {
		t.Errorf("Got code %q, want %q", code, wantCode)
	}
}

func TestToggleQuotePaste(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`v0 = $edit:insert:quote-paste`,
		`edit:toggle-quote-paste`,
		`v1 = $edit:insert:quote-paste`,
		`edit:toggle-quote-paste`,
		`v2 = $edit:insert:quote-paste`)

	v0 := getGlobal(f.Evaler, "v0").(bool)
	v1 := getGlobal(f.Evaler, "v1").(bool)
	v2 := getGlobal(f.Evaler, "v2").(bool)
	if v1 == v0 {
		t.Errorf("got v1 = v0")
	}
	if v2 == v1 {
		t.Errorf("got v2 = v1")
	}
}

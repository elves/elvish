package cliedit

import (
	"testing"
)

func TestBeforeReadline(t *testing.T) {
	ed, _, ev, cleanup := setup()
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

func TestAfterReadline(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `called = 0`)
	evalf(ev, `called-with = ''`)
	evalf(ev, `edit:after-readline = [
	             [code]{ called = (+ $called 1); called-with = $code } ]`)

	codeCh, _, _ := start(ed)
	feedInput(ttyCtrl, "test code\n")
	<-codeCh

	// TODO(xiaq): Test more precisely when after-readline is called.
	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
	if calledWith := ev.Global["called-with"].Get(); calledWith != "test code" {
		t.Errorf("called = %q, want %q", calledWith, "test code")
	}
}

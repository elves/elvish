package cliedit

import (
	"testing"
)

func TestBeforeReadline(t *testing.T) {
	ttyCtrl, ev, cleanup := setupWithRC(
		`called = 0`,
		`edit:before-readline = [ { called = (+ $called 1) } ]`)
	defer cleanup()

	// Wait for UI to stablize so that we can be sure that before-readline hooks
	// have been called.
	wantBufStable := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStable)

	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
}

func TestAfterReadline(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
	defer cleanup()
	evalf(ev, `called = 0`)
	evalf(ev, `called-with = ''`)
	evalf(ev, `edit:after-readline = [
	             [code]{ called = (+ $called 1); called-with = $code } ]`)

	codeCh, _, _ := start(ed)
	// Wait for UI to stablize so that we can be sure that after-readline hooks
	// are *not* called.
	wantBufStable := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStable)
	if called := ev.Global["called"].Get(); called != "0" {
		t.Errorf("called = %v, want 0", called)
	}

	// Input "test code", press Enter and wait until the editor is done.
	feedInput(ttyCtrl, "test code\n")
	<-codeCh

	if called := ev.Global["called"].Get(); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
	if calledWith := ev.Global["called-with"].Get(); calledWith != "test code" {
		t.Errorf("called = %q, want %q", calledWith, "test code")
	}
}

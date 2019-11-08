package cliedit

import (
	"testing"
)

func TestBeforeReadline(t *testing.T) {
	f := setupWithRC(
		`called = 0`,
		`edit:before-readline = [ { called = (+ $called 1) } ]`)
	defer f.Cleanup()

	// Wait for UI to stablize so that we can be sure that before-readline hooks
	// have been called.
	wantBufStable := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStable)

	if called := getGlobal(f.Evaler, "called"); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
}

func TestAfterReadline(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	evals(f.Evaler,
		`called = 0`,
		`called-with = ''`,
		`edit:after-readline = [
	             [code]{ called = (+ $called 1); called-with = $code } ]`)

	// Wait for UI to stablize so that we can be sure that after-readline hooks
	// are *not* called.
	wantBufStable := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStable)
	if called := getGlobal(f.Evaler, "called"); called != "0" {
		t.Errorf("called = %v, want 0", called)
	}

	// Input "test code", press Enter and wait until the editor is done.
	feedInput(f.TTYCtrl, "test code\n")
	f.Wait()

	if called := getGlobal(f.Evaler, "called"); called != 1.0 {
		t.Errorf("called = %v, want 1", called)
	}
	if calledWith := getGlobal(f.Evaler, "called-with"); calledWith != "test code" {
		t.Errorf("called = %q, want %q", calledWith, "test code")
	}
}

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

	testGlobal(t, f.Evaler, "called", 1.0)
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
	testGlobal(t, f.Evaler, "called", "0")

	// Input "test code", press Enter and wait until the editor is done.
	feedInput(f.TTYCtrl, "test code\n")
	f.Wait()

	testGlobals(t, f.Evaler, map[string]interface{}{
		"called":      1.0,
		"called-with": "test code",
	})
}

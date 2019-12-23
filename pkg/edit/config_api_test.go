package edit

import (
	"testing"

	"github.com/elves/elvish/pkg/cli/term"
)

func TestBeforeReadline(t *testing.T) {
	f := setup(rc(
		`called = 0`,
		`edit:before-readline = [ { called = (+ $called 1) } ]`))
	defer f.Cleanup()

	// Wait for UI to stablize so that we can be sure that before-readline hooks
	// have been called.
	f.TestTTY(t, "~> ", term.DotHere)

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
	f.TestTTY(t, "~> ", term.DotHere)
	testGlobal(t, f.Evaler, "called", "0")

	// Input "test code", press Enter and wait until the editor is done.
	feedInput(f.TTYCtrl, "test code\n")
	f.Wait()

	testGlobals(t, f.Evaler, map[string]interface{}{
		"called":      1.0,
		"called-with": "test code",
	})
}

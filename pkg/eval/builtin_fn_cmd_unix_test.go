//go:build unix

package eval_test

import (
	"os/exec"
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestHasExternal(t *testing.T) {
	Test(t,
		That("has-external sh").Puts(true),
		That("has-external random-invalid-command").Puts(false),
	)
}

func TestSearchExternal(t *testing.T) {
	Test(t,
		// Even on Unix systems we can't assume that commands like `sh` or
		// `test` are in a specific directory. Those commands might be in /bin
		// or /usr/bin. However, on all systems we currently support it will
		// be in /bin and, possibly, /usr/bin. So ensure we limit the search
		// to the one universal Unix directory for basic commands.
		That("{ tmp E:PATH = /bin;  search-external sh }").Puts("/bin/sh"),
		// We should check for a specific error if the external command cannot
		// be found. However, the current implementation of `search-external`
		// returns the raw error returned by a Go runtime function over which
		// we have no control.
		//
		// TODO: Replace the raw Go runtime `exec.LookPath` error with an
		// Elvish error; possibly wrapping the Go runtime error. Then tighten
		// this test to that specific error.
		That("search-external random-invalid-command").Throws(ErrorWithType(&exec.Error{})),
	)
}

func TestExternal(t *testing.T) {
	Test(t,
		That(`(external sh) -c 'echo external-sh'`).Prints("external-sh\n"),
	)
}

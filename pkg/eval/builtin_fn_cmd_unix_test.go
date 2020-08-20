// +build !windows,!plan9,!js

package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/env"
	"github.com/elves/elvish/pkg/util"
)

// Tests of the `builtin:has-external` command.
func TestBuiltinHasExternal(t *testing.T) {
	Test(t,
		That("has-external sh").Puts(true),
		That("has-external random-invalid-command").Puts(false),
	)
}

// Tests of the `builtin:search-external` command.
func TestBuiltinSearchExternal(t *testing.T) {
	tmpDir, cleanup := util.InTestDir()
	defer cleanup()

	mustWriteFile("sh", []byte("#!/bin/sh"), 0o755)
	mustWriteFile("mycmd", []byte("#!/bin/sh"), 0o755)

	// Even on UNIX systems we can't assume that commands like `sh` are in a
	// specific directory. Those commands might be in /bin or /usr/bin.
	// However, on all systems we currently support it will be in /bin and,
	// possibly, /usr/bin. So ensure we limit the search to the one universal
	// UNIX directory for basic commands.
	oldPath := os.Getenv(env.PATH)
	defer os.Setenv(env.PATH, oldPath)
	os.Setenv(env.PATH, "/bin:"+tmpDir)

	Test(t,
		That("search-external sh").Puts("/bin/sh"),
		// We should check for a specific error if the external command cannot
		// be found. However, the current implementation of `search-external`
		// returns the raw error returned by a Go runtime function over which
		// we have no control.
		//
		// TODO: Replace the raw Go runtime `exec.LookPath` error with an
		// Elvish error; possibly wrapping the Go runtime error. Then tighten
		// this test to that specific error.
		That("search-external invalid-command").ThrowsAny(),

		// Verify before of &all option.
		That("search-external &all sh").Puts("/bin/sh", filepath.Join(tmpDir, "sh")),
		That("search-external &all mycmd").Puts(filepath.Join(tmpDir, "mycmd")),
		That("search-external &all invalid-command").ThrowsMessage(
			"invalid-command: executable file not found in $E:PATH"),
	)
}

// Tests of the `builtin:external` command.
func TestBuiltinExternal(t *testing.T) {
	Test(t,
		That(`(external sh) -c 'echo external-sh'`).Prints("external-sh\n"),
	)
}

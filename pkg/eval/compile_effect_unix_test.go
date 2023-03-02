//go:build unix

package eval_test

import (
	"os"
	"strings"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestPipeline_ReaderGone_Unix(t *testing.T) {
	Test(t,
		// External commands terminated by SIGPIPE due to reader exiting early
		// raise ReaderGone, which is then suppressed.
		That("yes | true").DoesNothing(),
		That(
			"var reached = $false",
			"{ yes; reached = $true } | true",
			"put $reached",
		).Puts(false),
	)
}

func TestCommand_External(t *testing.T) {
	d := testutil.InTempDir(t)

	mustWriteScript("foo", "#!/bin/sh", "echo foo")
	mustWriteScript("lorem/ipsum", "#!/bin/sh", "echo lorem ipsum")

	testutil.Setenv(t, "PATH", d+"/bin")
	mustWriteScript("bin/hello", "#!/bin/sh", "echo hello")

	Test(t,
		// External commands, searched and relative
		That("hello").Prints("hello\n"),
		That("./foo").Prints("foo\n"),
		That("lorem/ipsum").Prints("lorem ipsum\n"),
		// Using the explicit e: namespace.
		That("e:hello").Prints("hello\n"),
		That("e:./foo").Prints("foo\n"),
		// Relative external commands may be a dynamic string.
		That("var x = ipsum", "lorem/$x").Prints("lorem ipsum\n"),
		// Searched external commands may not be a dynamic string.
		That("var x = hello; $x").Throws(
			errs.BadValue{What: "command",
				Valid: "callable or string containing slash", Actual: "hello"},
			"$x"),

		// Using new FD as destination in external commands.
		// Regression test against b.elv.sh/788.
		That("./foo 5</dev/null").Prints("foo\n"),

		// Using pragma to allow or disallow implicit searched commands
		That("pragma unknown-command = disallow", "hello").DoesNotCompile("unknown command disallowed by current pragma"),
		That("pragma unknown-command = external", "hello").Prints("hello\n"),
		// Pragma applies to subscope
		That("pragma unknown-command = disallow", "{ hello }").DoesNotCompile("unknown command disallowed by current pragma"),
		// Explicit uses with e: is always allowed
		That("pragma unknown-command = disallow", "e:hello").Prints("hello\n"),
		// Relative external commands are always allowed
		That("pragma unknown-command = disallow", "./foo").Prints("foo\n"),
		That("pragma unknown-command = disallow", "var x = ./foo", "$x").Prints("foo\n"),
	)
}

func mustWriteScript(name string, lines ...string) {
	must.WriteFile(name, strings.Join(lines, "\n"))
	must.OK(os.Chmod(name, 0700))
}

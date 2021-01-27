// +build !windows,!plan9

package eval_test

import (
	"path/filepath"
	"strings"
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestCompileEffectUnix(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	mustWriteScript("foo", "#!/bin/sh", "echo foo")
	mustWriteScript("lorem/ipsum", "#!/bin/sh", "echo lorem ipsum")

	Test(t,
		// External commands.
		That("./foo").Prints("foo\n"),
		That("lorem/ipsum").Prints("lorem ipsum\n"),
		// Using the explicit e: namespace.
		That("e:./foo").Prints("foo\n"),
		// Names of external commands may be built dynamically.
		That("x = ipsum", "lorem/$x").Prints("lorem ipsum\n"),
		// Using new FD as destination in external commands.
		// Regression test against b.elv.sh/788.
		That("./foo 5</dev/null").Prints("foo\n"),
	)
}

func mustWriteScript(name string, lines ...string) {
	testutil.MustMkdirAll(filepath.Dir(name))
	testutil.MustWriteFile(name, []byte(strings.Join(lines, "\n")), 0700)
}

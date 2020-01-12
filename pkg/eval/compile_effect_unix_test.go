// +build !windows,!plan9

package eval

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestCompileEffectUnix(t *testing.T) {
	_, cleanup := util.InTestDir()
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
	MustMkdirAll(filepath.Dir(name), 0700)
	MustWriteFile(name, []byte(strings.Join(lines, "\n")), 0700)
}

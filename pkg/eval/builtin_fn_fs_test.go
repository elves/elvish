package eval

import (
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/parse"
)

func TestBuiltinFnFS(t *testing.T) {
	tmpHome, cleanup := InTempHome()
	defer cleanup()

	mustMkdirAll("dir", 0700)
	mustCreateEmpty("file")

	Test(t,
		That(`path-base a/b/c.png`).Puts("c.png"),
		That("tilde-abbr "+parse.Quote(filepath.Join(tmpHome, "foobar"))).
			Puts(filepath.Join("~", "foobar")),

		That(`-is-dir ~/dir`).Puts(true),
		That(`-is-dir ~/file`).Puts(false),
	)
}

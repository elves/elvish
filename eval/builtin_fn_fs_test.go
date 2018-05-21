package eval

import (
	"path/filepath"
	"testing"

	"github.com/elves/elvish/parse"
)

func TestBuiltinFnFS(t *testing.T) {
	pathSep := string(filepath.Separator)
	InTempHome(func(tmpHome string) {
		MustMkdirAll("dir", 0700)
		MustCreateEmpty("file")

		Test(t, []TestCase{
			That(`path-base a/b/c.png`).Puts("c.png"),
			That("tilde-abbr " + parse.Quote(filepath.Join(tmpHome, "foobar"))).Puts(
				"~" + pathSep + "foobar"),

			That(`-is-dir ~/dir`).Puts(true),
			That(`-is-dir ~/file`).Puts(false),
		})
	})
}

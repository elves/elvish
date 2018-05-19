package eval

import (
	"path/filepath"
	"testing"
)

func TestBuiltinFnFS(t *testing.T) {
	pathSep := string(filepath.Separator)
	test(t, []TestCase{
		That(`path-base a/b/c.png`).Puts("c.png"),
		That(`tilde-abbr $E:HOME'` + pathSep + `'foobar`).Puts(
			"~" + pathSep + "foobar"),

		// see testmain_test.go for setup
		That(`-is-dir ~/dir`).Puts(true),
		That(`-is-dir ~/lorem`).Puts(false),
	})
}

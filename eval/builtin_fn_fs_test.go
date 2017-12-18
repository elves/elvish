package eval

import (
	"path/filepath"
)

var pathSep = string(filepath.Separator)

func init() {
	addToEvalTests([]evalTest{
		{`path-base a/b/c.png`, want{out: strs("c.png")}},
		{`tilde-abbr $E:HOME` + pathSep + `foobar`,
			want{out: strs("~" + pathSep + "foobar")}},

		{`-is-dir ~/dir`, wantTrue}, // see testmain_test.go for setup
		{`-is-dir ~/lorem`, wantFalse},
	})
}

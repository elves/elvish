package path

import (
	"path/filepath"
	"regexp"
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

var testDir = testutil.Dir{
	"d": testutil.Dir{
		"f": "",
	},
}

// A regular expression fragment to match the directory part of an absolute
// path. QuoteMeta is needed since on Windows filepath.Separator is '\\'.
var anyDir = "^.*" + regexp.QuoteMeta(string(filepath.Separator))

func TestPath(t *testing.T) {
	testutil.InTempDir(t)
	testutil.ApplyDir(testDir)

	absPath := must.OK1(filepath.Abs("a/b/c.png"))

	TestWithEvalerSetup(t, Use("path", Ns, "file", file.Ns),
		//  All the functions in path: are either simple wrappers of Go
		//  functions or compatibility aliases of their os: counterparts.
		//
		// As a result, the tests are just simple "smoke tests" to ensure that
		// they exist and map to the correct function.
		That("put $path:list-separator").Puts(string(filepath.ListSeparator)),
		That("put $path:separator").Puts(string(filepath.Separator)),
		That("path:abs a/b/c.png").Puts(absPath),
		That("path:base a/b/d.png").Puts("d.png"),
		That("path:clean ././x").Puts("x"),
		That("path:clean a/b/.././c").Puts(filepath.Join("a", "c")),
		That("path:dir a/b/d.png").Puts(filepath.Join("a", "b")),
		That("path:ext a/b/e.png").Puts(".png"),
		That("path:ext a/b/s").Puts(""),
		That("path:is-abs a/b/s").Puts(false),
		That("path:is-abs "+absPath).Puts(true),
		That("path:join a b c").Puts(filepath.Join("a", "b", "c")),
		// Compatibility aliases.
		That("path:eval-symlinks d").Puts("d"),
		That("path:is-dir d").Puts(true),
		That("path:is-regular d/f").Puts(true),
		That("var x = (path:temp-dir)", "rmdir $x", "put $x").Puts(
			StringMatching(anyDir+`elvish-.*$`)),
		That("var f = (path:temp-file)", "put $f[name]", "file:close $f", "rm $f[name]").
			Puts(StringMatching(anyDir+`elvish-.*$`)),
	)
}

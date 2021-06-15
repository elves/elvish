package path

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
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
	tmpdir, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.ApplyDir(testDir)

	absPath, err := filepath.Abs("a/b/c.png")
	if err != nil {
		panic("unable to convert a/b/c.png to an absolute path")
	}

	TestWithSetup(t, importPathModule,
		// This block of tests is not meant to be comprehensive. Their primary purpose is to simply
		// ensure the Elvish command is correctly mapped to the relevant Go function. We assume the
		// Go function behaves correctly.
		That("path:abs a/b/c.png").Puts(absPath),
		That("path:base a/b/d.png").Puts("d.png"),
		That("path:clean ././x").Puts("x"),
		That("path:clean a/b/.././c").Puts(filepath.Join("a", "c")),
		That("path:dir a/b/d.png").Puts(filepath.Join("a", "b")),
		That("path:ext a/b/e.png").Puts(".png"),
		That("path:ext a/b/s").Puts(""),
		That("path:is-abs a/b/s").Puts(false),
		That("path:is-abs "+absPath).Puts(true),

		// Elvish "path:" module functions that are not trivial wrappers around a Go stdlib function
		// should have comprehensive tests below this comment.
		That("path:is-dir "+tmpdir).Puts(true),
		That("path:is-dir d").Puts(true),
		That("path:is-dir d/f").Puts(false),
		That("path:is-dir bad").Puts(false),

		That("path:is-regular "+tmpdir).Puts(false),
		That("path:is-regular d").Puts(false),
		That("path:is-regular d/f").Puts(true),
		That("path:is-regular bad").Puts(false),

		// Verify the commands for creating temporary filesystem objects work correctly.
		That("x = (path:temp-dir)", "rmdir $x", "put $x").Puts(
			MatchingRegexp{Pattern: anyDir + `elvish-.*$`}),
		That("x = (path:temp-dir 'x-*.y')", "rmdir $x", "put $x").Puts(
			MatchingRegexp{Pattern: anyDir + `x-.*\.y$`}),
		That("x = (path:temp-dir &dir=. 'x-*.y')", "rmdir $x", "put $x").Puts(
			MatchingRegexp{Pattern: `^x-.*\.y$`}),
		That("x = (path:temp-dir &dir=.)", "rmdir $x", "put $x").Puts(
			MatchingRegexp{Pattern: `^elvish-.*$`}),
		That("path:temp-dir a b").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"path:temp-dir a b"),

		That("f = (path:temp-file)", "fclose $f", "put $f[fd]", "rm $f[name]").
			Puts(-1),
		That("f = (path:temp-file)", "put $f[name]", "fclose $f", "rm $f[name]").
			Puts(MatchingRegexp{Pattern: anyDir + `elvish-.*$`}),
		That("f = (path:temp-file 'x-*.y')", "put $f[name]", "fclose $f", "rm $f[name]").
			Puts(MatchingRegexp{Pattern: anyDir + `x-.*\.y$`}),
		That("f = (path:temp-file &dir=. 'x-*.y')", "put $f[name]", "fclose $f", "rm $f[name]").
			Puts(MatchingRegexp{Pattern: `^x-.*\.y$`}),
		That("f = (path:temp-file &dir=.)", "put $f[name]", "fclose $f", "rm $f[name]").
			Puts(MatchingRegexp{Pattern: `^elvish-.*$`}),
		That("path:temp-file a b").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"path:temp-file a b"),
	)
}

var symlinks = []struct {
	path   string
	target string
}{
	{"d/s-f", "f"},
	{"s-d", "d"},
	{"s-d-f", "d/f"},
	{"s-bad", "bad"},
}

func TestPath_Symlink(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.ApplyDir(testDir)
	// testutil.ApplyDir(testDirSymlinks)
	for _, link := range symlinks {
		err := os.Symlink(link.target, link.path)
		if err != nil {
			// Creating symlinks requires a special permission on Windows. If
			// the user doesn't have that permission, just skip the whole test.
			t.Skip(err)
		}
	}

	TestWithSetup(t, importPathModule,
		That("path:eval-symlinks d/f").Puts(filepath.Join("d", "f")),
		That("path:eval-symlinks d/s-f").Puts(filepath.Join("d", "f")),
		That("path:eval-symlinks s-d/f").Puts(filepath.Join("d", "f")),
		That("path:eval-symlinks s-bad").Throws(AnyError),

		That("path:is-dir s-d").Puts(false),
		That("path:is-dir s-d &follow-symlink").Puts(true),
		That("path:is-dir s-d-f").Puts(false),
		That("path:is-dir s-d-f &follow-symlink").Puts(false),
		That("path:is-dir s-bad").Puts(false),
		That("path:is-dir s-bad &follow-symlink").Puts(false),
		That("path:is-dir bad").Puts(false),
		That("path:is-dir bad &follow-symlink").Puts(false),

		That("path:is-regular s-d").Puts(false),
		That("path:is-regular s-d &follow-symlink").Puts(false),
		That("path:is-regular s-d-f").Puts(false),
		That("path:is-regular s-d-f &follow-symlink").Puts(true),
		That("path:is-regular s-bad").Puts(false),
		That("path:is-regular s-bad &follow-symlink").Puts(false),
		That("path:is-regular bad").Puts(false),
		That("path:is-regular bad &follow-symlink").Puts(false),
	)
}

func importPathModule(ev *eval.Evaler) {
	ev.AddGlobal(eval.NsBuilder{}.AddNs("path", Ns).Ns())
}

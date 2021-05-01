package path

import (
	"path/filepath"
	"regexp"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

var testDir = testutil.Dir{
	"d1": testutil.Dir{
		"f": testutil.Symlink{Target: filepath.Join("d2", "f")},
		"d2": testutil.Dir{
			"empty": "",
			"f":     "",
			"g":     testutil.Symlink{Target: "f"},
		},
	},
	"s1": testutil.Symlink{Target: filepath.Join("d1", "d2")},
}

func TestPath(t *testing.T) {
	tmpdir, cleanup := testutil.InTestDir()
	defer cleanup()
	testutil.ApplyDir(testDir)

	absPath, err := filepath.Abs("a/b/c.png")
	if err != nil {
		panic("unable to convert a/b/c.png to an absolute path")
	}

	// This is needed for path tests that use a regexp for validating a path since Windows uses a
	// backslash as the path separator and a backslash is special in a regexp.
	sep := regexp.QuoteMeta(string(filepath.Separator))

	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("path", Ns).Ns())
	}
	TestWithSetup(t, setup,
		// This block of tests is not meant to be comprehensive. Their primary purpose is to simply
		// ensure the Elvish command is correctly mapped to the relevant Go function. We assume the
		// Go function behaves correctly.
		That(`path:abs a/b/c.png`).Puts(absPath),
		That(`path:base a/b/d.png`).Puts("d.png"),
		That(`path:clean ././x`).Puts("x"),
		That(`path:clean a/b/.././c`).Puts(filepath.Join("a", "c")),
		That(`path:dir a/b/d.png`).Puts(filepath.Join("a", "b")),
		That(`path:ext a/b/e.png`).Puts(".png"),
		That(`path:ext a/b/s`).Puts(""),
		That(`path:is-abs a/b/s`).Puts(false),
		That(`path:is-abs `+absPath).Puts(true),
		That(`path:eval-symlinks d1/d2`).Puts(filepath.Join("d1", "d2")),
		That(`path:eval-symlinks d1/d2/f`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1`).Puts(filepath.Join("d1", "d2")),
		That(`path:eval-symlinks d1/f`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1/g`).Puts(filepath.Join("d1", "d2", "f")),
		That(`path:eval-symlinks s1/empty`).Puts(filepath.Join("d1", "d2", "empty")),

		// Elvish `path:` module functions that are not trivial wrappers around a Go stdlib function
		// should have comprehensive tests below this comment.
		That(`path:is-dir a/b/s`).Puts(false),
		That(`path:is-dir `+tmpdir).Puts(true),
		That(`path:is-dir s1`).Puts(false),
		That(`path:is-regular a/b/s`).Puts(false),
		That(`path:is-regular `+tmpdir).Puts(false),
		That(`path:is-regular d1/f`).Puts(false),
		That(`path:is-regular d1/d2/f`).Puts(true),
		That(`path:is-regular s1/f`).Puts(true),

		// Verify the commands for creating temporary filesystem objects work correctly.
		That(`x = (path:temp-dir)`, `rmdir $x`, `put $x`).Puts(
			MatchingRegexp{Pattern: `^.*` + sep + `elvish-.*$`}),
		That(`x = (path:temp-dir 'x-*.y')`, `rmdir $x`, `put $x`).Puts(
			MatchingRegexp{Pattern: `^.*` + sep + `x-.*\.y$`}),
		That(`x = (path:temp-dir &dir=. 'x-*.y')`, `rmdir $x`, `put $x`).Puts(
			MatchingRegexp{Pattern: `^x-.*\.y$`}),
		That(`x = (path:temp-dir &dir=.)`, `rmdir $x`, `put $x`).Puts(
			MatchingRegexp{Pattern: `^elvish-.*$`}),
		That(`path:temp-dir a b`).Throws(
			errs.ArityMismatch{What: "arguments here", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"path:temp-dir a b"),

		That(`f = (path:temp-file)`, `fclose $f`, `put $f[fd]`, `rm $f[name]`).
			Puts(-1),
		That(`f = (path:temp-file)`, `put $f[name]`, `fclose $f`, `rm $f[name]`).
			Puts(MatchingRegexp{Pattern: `^.*` + sep + `elvish-.*$`}),
		That(`f = (path:temp-file 'x-*.y')`, `put $f[name]`, `fclose $f`, `rm $f[name]`).
			Puts(MatchingRegexp{Pattern: `^.*` + sep + `x-.*\.y$`}),
		That(`f = (path:temp-file &dir=. 'x-*.y')`, `put $f[name]`, `fclose $f`, `rm $f[name]`).
			Puts(MatchingRegexp{Pattern: `^x-.*\.y$`}),
		That(`f = (path:temp-file &dir=.)`, `put $f[name]`, `fclose $f`, `rm $f[name]`).
			Puts(MatchingRegexp{Pattern: `^elvish-.*$`}),
		That(`path:temp-file a b`).Throws(
			errs.ArityMismatch{What: "arguments here", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"path:temp-file a b"),
	)
}

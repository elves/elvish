package os_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/mods/file"
	osmod "src.elv.sh/pkg/mods/os"
	"src.elv.sh/pkg/testutil"
)

func TestFSModifications(t *testing.T) {
	// Also tests -is-exists and -is-not-exists
	TestWithSetup(t, func(t *testing.T, ev *eval.Evaler) {
		testutil.InTempDir(t)
		useOS(ev)
	},
		// mkdir
		That(`os:mkdir d; os:is-dir d`).Puts(true),
		That(`os:mkdir d; try { os:mkdir d } catch e { os:-is-exist $e }`).
			Puts(true),

		// remove
		That(`echo > f; os:exists f; os:remove f; os:exists f`).
			Puts(true, false),
		That(`os:mkdir d; os:exists d; os:remove d; os:exists d`).
			Puts(true, false),
		That(`os:mkdir d; echo > d/file; os:remove d`).
			Throws(ErrorWithType(&fs.PathError{})),
		That(`try { os:remove d } catch e { os:-is-not-exist $e }`).
			Puts(true),
		That(`os:remove ""`).Throws(osmod.ErrEmptyPath),

		// remove-all
		That(`os:mkdir d; echo > d/file; os:remove-all d; os:exists d`).
			Puts(false),
		That(`os:mkdir d; echo > d/file; os:remove-all $pwd/d; os:exists d`).
			Puts(false),
		That(`os:remove-all d`).DoesNothing(),
		That(`os:remove-all ""`).Throws(osmod.ErrEmptyPath),

		// The success cases for chmod are in os_unix_test.go, since they depend
		// on the exact value of the perm bits and special modes.
		//
		// TODO: Add tests for Windows after Elvish supports bitwise operations.

		// chmod errors
		That(`os:chmod -1 d`).
			Throws(errs.OutOfRange{What: "permission bits",
				ValidLow: "0", ValidHigh: "0o777", Actual: "-1"}),
		// TODO: This error should be more informative and point out that it is
		// the special modes that should be iterable
		That(`os:chmod &special-modes=(num 0) 0 d`).
			Throws(ErrorWithMessage("cannot iterate number")),
		That(`os:chmod &special-modes=[bad] 0 d`).
			Throws(errs.BadValue{What: "special mode",
				Valid: "setuid, setgid or sticky", Actual: "bad"}),
	)
}

var testDir = Dir{
	"d": Dir{
		"f": "",
	},
}

func TestFilePredicates(t *testing.T) {
	tmpdir := InTempDir(t)
	ApplyDir(testDir)

	TestWithEvalerSetup(t, useOS,
		That("os:exists "+tmpdir).Puts(true),
		That("os:exists d").Puts(true),
		That("os:exists d/f").Puts(true),
		That("os:exists bad").Puts(false),

		That("os:is-dir "+tmpdir).Puts(true),
		That("os:is-dir d").Puts(true),
		That("os:is-dir d/f").Puts(false),
		That("os:is-dir bad").Puts(false),

		That("os:is-regular "+tmpdir).Puts(false),
		That("os:is-regular d").Puts(false),
		That("os:is-regular d/f").Puts(true),
		That("os:is-regular bad").Puts(false),
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

func TestFilePredicates_Symlinks(t *testing.T) {
	InTempDir(t)
	ApplyDir(testDir)
	for _, link := range symlinks {
		err := os.Symlink(link.target, link.path)
		if err != nil {
			// Creating symlinks requires a special permission on Windows. If
			// the user doesn't have that permission, just skip the whole test.
			t.Skip(err)
		}
	}

	TestWithEvalerSetup(t, Use("os", osmod.Ns),
		That("os:eval-symlinks d/f").Puts(filepath.Join("d", "f")),
		That("os:eval-symlinks d/s-f").Puts(filepath.Join("d", "f")),
		That("os:eval-symlinks s-d/f").Puts(filepath.Join("d", "f")),
		That("os:eval-symlinks s-bad").Throws(ErrorWithType(&os.PathError{})),

		That("os:exists s-d").Puts(true),
		That("os:exists s-d &follow-symlink").Puts(true),
		That("os:exists s-d-f").Puts(true),
		That("os:exists s-d-f &follow-symlink").Puts(true),
		That("os:exists s-bad").Puts(true),
		That("os:exists s-bad &follow-symlink").Puts(false),
		That("os:exists bad").Puts(false),
		That("os:exists bad &follow-symlink").Puts(false),

		That("os:is-dir s-d").Puts(false),
		That("os:is-dir s-d &follow-symlink").Puts(true),
		That("os:is-dir s-d-f").Puts(false),
		That("os:is-dir s-d-f &follow-symlink").Puts(false),
		That("os:is-dir s-bad").Puts(false),
		That("os:is-dir s-bad &follow-symlink").Puts(false),
		That("os:is-dir bad").Puts(false),
		That("os:is-dir bad &follow-symlink").Puts(false),

		That("os:is-regular s-d").Puts(false),
		That("os:is-regular s-d &follow-symlink").Puts(false),
		That("os:is-regular s-d-f").Puts(false),
		That("os:is-regular s-d-f &follow-symlink").Puts(true),
		That("os:is-regular s-bad").Puts(false),
		That("os:is-regular s-bad &follow-symlink").Puts(false),
		That("os:is-regular bad").Puts(false),
		That("os:is-regular bad &follow-symlink").Puts(false),
	)
}

// A regular expression fragment to match the directory part of an absolute
// path. QuoteMeta is needed since on Windows filepath.Separator is '\\'.
var anyDir = "^.*" + regexp.QuoteMeta(string(filepath.Separator))

func TestTempDirFile(t *testing.T) {
	InTempDir(t)

	TestWithEvalerSetup(t, Use("os", osmod.Ns, "file", file.Ns),
		That("var x = (os:temp-dir)", "rmdir $x", "put $x").Puts(
			StringMatching(anyDir+`elvish-.*$`)),
		That("var x = (os:temp-dir 'x-*.y')", "rmdir $x", "put $x").Puts(
			StringMatching(anyDir+`x-.*\.y$`)),
		That("var x = (os:temp-dir &dir=. 'x-*.y')", "rmdir $x", "put $x").Puts(
			StringMatching(`^(\.[/\\])?x-.*\.y$`)),
		That("var x = (os:temp-dir &dir=.)", "rmdir $x", "put $x").Puts(
			StringMatching(`^(\.[/\\])?elvish-.*$`)),
		That("os:temp-dir a b").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"os:temp-dir a b"),

		That("var f = (os:temp-file)", "file:close $f", "put $f[fd]", "rm $f[name]").
			Puts(-1),
		That("var f = (os:temp-file)", "put $f[name]", "file:close $f", "rm $f[name]").
			Puts(StringMatching(anyDir+`elvish-.*$`)),
		That("var f = (os:temp-file 'x-*.y')", "put $f[name]", "file:close $f", "rm $f[name]").
			Puts(StringMatching(anyDir+`x-.*\.y$`)),
		That("var f = (os:temp-file &dir=. 'x-*.y')", "put $f[name]", "file:close $f", "rm $f[name]").
			Puts(StringMatching(`^(\.[/\\])?x-.*\.y$`)),
		That("var f = (os:temp-file &dir=.)", "put $f[name]", "file:close $f", "rm $f[name]").
			Puts(StringMatching(`^(\.[/\\])?elvish-.*$`)),
		That("os:temp-file a b").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: 2},
			"os:temp-file a b"),
	)
}

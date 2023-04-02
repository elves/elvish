package os_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	os_mod "src.elv.sh/pkg/mods/os"
	"src.elv.sh/pkg/mods/path"
	"src.elv.sh/pkg/mods/unix"
	"src.elv.sh/pkg/testutil"
)

var (
	ErrorWithType        = evaltest.ErrorWithType
	ErrorWithMessage     = evaltest.ErrorWithMessage
	TestWithSetup        = evaltest.TestWithEvalerSetup
	That                 = evaltest.That
	StringMatchingRegexp = evaltest.StringMatching
)

var testDir = testutil.Dir{
	"d": testutil.Dir{
		"f": "",
	},
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

func TestOS(t *testing.T) {
	tmpdir := testutil.InTempDir(t)
	testutil.ApplyDir(testDir)
	for _, link := range symlinks {
		err := os.Symlink(link.target, link.path)
		if err != nil {
			// Creating symlinks requires a special permission on Windows. If
			// the user doesn't have that permission, just skip the whole test.
			t.Skip(err)
		}
	}
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("os", os_mod.Ns))
		ev.ExtendGlobal(eval.BuildNs().AddNs("path", path.Ns))
		ev.ExtendGlobal(eval.BuildNs().AddNs("unix", unix.Ns))
	}
	// TODO: Run every test case in a new temp directory so that we don't have
	// to keep track of which filenames have been used in previous test cases.
	TestWithSetup(t, setup,
		// -is-exist and -is-not-exist are tested in test cases of other
		// commands

		// mkdir
		That(`os:mkdir d1; path:is-dir d1`).Puts(true),
		That(`os:mkdir d2; try { os:mkdir d2 } catch e { os:-is-exist $e }`).
			Puts(true),

		// remove
		That(`echo > f1; os:exists f1; os:remove f1; os:exists f1`).
			Puts(true, false),
		That(`os:mkdir d3; os:exists d3; os:remove d3; os:exists d3`).
			Puts(true, false),
		That(`os:mkdir d4; echo > d4/file; os:remove d4`).
			Throws(ErrorWithType(&fs.PathError{})),
		That(`try { os:remove d5 } catch e { os:-is-not-exist $e }`).
			Puts(true),
		That(`os:remove ""`).Throws(os_mod.ErrEmptyPath),

		// remove-all
		That(`os:mkdir d6; echo > d6/file; os:remove-all d6; os:exists d6`).
			Puts(false),
		That(`os:mkdir d7; echo > d7/file; os:remove-all $pwd/d7; os:exists d7`).
			Puts(false),
		That(`os:remove-all d8`).DoesNothing(),
		That(`os:remove-all ""`).Throws(os_mod.ErrEmptyPath),

		// stat
		That("put (os:stat d)[path]").Puts("d"),
		That("put (os:stat d/f)[path]").Puts("d/f"),
		That("put (os:stat d/f)[abs-path]").Puts(filepath.Join(tmpdir, "d", "f")),
		That("put (os:stat d/f)[size]").Puts(0),

		// stat symlinks
		That("put (os:stat s-d)[is-dir]").Puts(false),
		That("put (os:stat s-d)[is-regular]").Puts(false),
		That("put (os:stat &follow-symlink s-d)[is-dir]").Puts(true),
		That("put (os:stat s-d-f)[is-dir]").Puts(false),
		That("put (os:stat s-d-f)[is-regular]").Puts(false),
		That("put (os:stat &follow-symlink s-d-f)[is-regular]").Puts(true),
		That("put (os:stat s-d-f)[is-symlink]").Puts(true),
		That("put (os:stat &follow-symlink s-d-f)[is-symlink]").Puts(false),
	)

	if unix.ExposeUnixNs {
		testutil.Umask(t, 0)

		TestWithSetup(t, setup,
			// TODO: Remove the need for ls when there is os:lstat.
			That(`os:mkdir &perm=0o400 d400; ls -ld d400 | slurp`).
				Puts(StringMatchingRegexp(`dr--.*`)),

			// TODO: Remove the need for ln when there is os:symlink.
			That(`ln -s bad l; os:exists l; os:exists &follow-symlink=$true l`).
				Puts(true, false),
		)
	}
}

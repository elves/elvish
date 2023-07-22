package os_test

import (
	"io/fs"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/os"
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

func TestOS(t *testing.T) {
	testutil.InTempDir(t)
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("os", os.Ns))
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
		That(`os:remove ""`).Throws(os.ErrEmptyPath),

		// remove-all
		That(`os:mkdir d6; echo > d6/file; os:remove-all d6; os:exists d6`).
			Puts(false),
		That(`os:mkdir d7; echo > d7/file; os:remove-all $pwd/d7; os:exists d7`).
			Puts(false),
		That(`os:remove-all d8`).DoesNothing(),
		That(`os:remove-all ""`).Throws(os.ErrEmptyPath),
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

//go:build unix

package os_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/testutil"
)

func TestChmod(t *testing.T) {
	// Note: On FreeBSD the /tmp directory is usually mounted with the "suiddir"
	// option. This causes new directories, such as created by this unit test,
	// to inherit the /tmp group owner. If the account running this unit test is
	// not a member of that group (typically "wheel") you'll get test failures.
	TestWithSetup(t, func(t *testing.T, ev *eval.Evaler) {
		testutil.InTempDir(t)
		useOS(ev)
	},
		That(`os:mkdir d1; os:chmod 0o400 d1; put (os:stat d1)[perm]`).Puts(0o400),
		That(`os:mkdir d2; os:chmod &special-modes=[setuid setgid sticky] 0o400 d2; put (os:stat d2)[special-modes]`).
			Puts(vals.MakeList("setuid", "setgid", "sticky")),
	)
}

func TestMkdirPerm(t *testing.T) {
	testutil.Umask(t, 0)
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, useOS,
		// TODO: Remove the need for ls when there is os:stat.
		That(`os:mkdir &perm=0o400 d400; put (os:stat d400)[perm]`).
			Puts(0o400),
	)
}

func TestExists_BadSymlink(t *testing.T) {
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, useOS,
		// TODO: Remove the need for ln when there is os:symlink.
		That(`ln -s bad l; os:exists l; os:exists &follow-symlink=$true l`).
			Puts(true, false),
	)
}

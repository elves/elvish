//go:build unix

package os_test

import (
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestMkdirPerm(t *testing.T) {
	testutil.Umask(t, 0)
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, useOS,
		// TODO: Remove the need for ls when there is os:stat.
		That(`os:mkdir &perm=0o400 d400; ls -ld d400 | slurp`).
			Puts(StringMatching(`dr--.*`)),
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

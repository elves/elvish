//go:build unix

package os_test

import (
	"os"
	"runtime"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestChmod(t *testing.T) {
	TestWithSetup(t, func(t *testing.T, ev *eval.Evaler) {
		tmpdir := testutil.InTempDir(t)
		if runtime.GOOS == "freebsd" {
			// Work around a quirk of FreeBSD's mkdir
			// (https://github.com/golang/go/issues/19596).
			must.OK(os.Chown(tmpdir, os.Getuid(), os.Getgid()))
		}
		useOS(ev)
	},
		That(`os:mkdir d; os:chmod 0o400 d; put (os:stat d)[perm]`).Puts(0o400),
		That(`os:mkdir d; os:chmod &special-modes=[setuid setgid sticky] 0o400 d; put (os:stat d)[special-modes]`).
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

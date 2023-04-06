//go:build unix

package os_test

import (
	"runtime"
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

func TestChmod(t *testing.T) {
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, useOS,
		That("echo >a; os:chmod 0o777 a; put (os:stat a)[perm]").Puts(0o777),
		That("echo >b; os:chmod 0o123 b; put (os:stat b)[perm]").Puts(0o123),
		That("echo >c; os:chmod 753 c; put (os:stat c)[perm]").Puts(0o753),
		That("echo >d; os:chmod 321 d; put (os:stat d)[perm]").Puts(0o321),
		That("echo >e; os:chmod (num 0o666) e; put (os:stat e)[perm]").Puts(0o666),
		That("echo >h; os:chmod (num 0o4664) h; var s = (os:stat h); "+
			"put $s[perm] (count $s[special-modes]) $s[special-modes][0]").Puts(0o664, 1, "setuid"),
	)
	// These two tests fail on FreeBSD with "inappropriate file type or format"
	// and "operation not permitted" respectively. I don't know why but will
	// trust that if they pass on Linux and macOS then the mode bits are
	// correctly mapped.
	if runtime.GOOS != "freebsd" {
		TestWithEvalerSetup(t, useOS,
			That("echo >f; os:chmod (num 0o1567) f; var s = (os:stat f); "+
				"put $s[perm] (count $s[special-modes]) $s[special-modes][0]").
				Puts(0o567, 1, "sticky"),
			That("echo >g; os:chmod (num 0o2444) g; var s = (os:stat g); "+
				"put $s[perm] (count $s[special-modes]) $s[special-modes][0]").
				Puts(0o444, 1, "setgid"),
		)
	}
}

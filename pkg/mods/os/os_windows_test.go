//go:build windows

package os_test

import (
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestChmod(t *testing.T) {
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, useOS,
		// These tests are weird for someone expecting Unix semantics. On MS
		// Windows the read bit is always set, there is no execute bit, and only
		// the write bit can be changed. There is no distinction between
		// user/group/public permissions.
		That("echo >a; os:chmod 0o707 a; put (os:stat a)[perm]").Puts(0o666),
		That("echo >b; os:chmod 0o131 b; put (os:stat b)[perm]").Puts(0o444),
		That("echo >c; os:chmod 0o113 c; put (os:stat c)[perm]").Puts(0o444),
	)
}

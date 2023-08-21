//go:build unix

package os_test

import (
	"net"
	"testing"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/must"
)

func TestStat_Type_Unix(t *testing.T) {
	InTempDir(t)
	must.OK(unix.Mkfifo("fifo", 0o600))
	listener := must.OK1(net.Listen("unix", "./sock"))
	defer listener.Close()

	TestWithEvalerSetup(t, useOS,
		That(`os:stat fifo`).Puts(MapContainingPairs("type", "named-pipe")),
		That(`os:stat sock`).Puts(MapContainingPairs("type", "socket")),
		That(`os:stat /dev/null`).
			Puts(MapContainingPairs("type", "char-device")),
	)
}

func TestStat_Sys_Unix(t *testing.T) {
	InTempDir(t)
	ApplyDir(Dir{"file": "123456"})

	TestWithEvalerSetup(t, useOS,
		That(`os:stat file`).Puts(
			MapContainingPairs("sys", MapContainingPairs(
				"dev", AnyInteger,
				"ino", AnyInteger,
				"nlink", 1,
				"uid", AnyInteger,
				"gid", AnyInteger,
				"rdev", AnyInteger,
				"blksize", AnyInteger,
				"blocks", AnyInteger,
			))),
	)
}

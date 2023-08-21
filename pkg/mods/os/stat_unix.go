//go:build unix

package os

import (
	"syscall"

	"src.elv.sh/pkg/eval/vals"
)

var extraStatFields = map[string]func(*syscall.Stat_t) uint64{}

func statSysMap(sys any) vals.Map {
	st := sys.(*syscall.Stat_t)
	m := vals.MakeMap(
		"dev", stNum(st.Dev),
		"ino", stNum(st.Ino),
		"nlink", stNum(st.Nlink),
		"uid", stNum(st.Uid),
		"gid", stNum(st.Gid),
		"rdev", stNum(st.Rdev),
		// TODO: atim, mtim, ctim
		"blksize", stNum(st.Blksize),
		"blocks", stNum(st.Blocks),
	)
	for name, f := range extraStatFields {
		m = m.Assoc(name, stNum(f(st)))
	}
	return m
}

// Converts a stat_t field to Num. All of the these fields are non-negative even
// if they are signed, so we convert them to uint64 first.
func stNum[T interface {
	int16 | uint16 | int32 | uint32 | int64 | uint64
}](x T) vals.Num {
	return vals.Uint64ToNum(uint64(x))
}

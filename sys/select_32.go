// +build darwin 386,linux arm,linux mips,linux mipsle,linux

// The type of FdSet.Bits is different on different platforms.
// This file is for those where FdSet.Bits is []int32.

package sys

const NFDBits = 32

func index(fd int) (idx uint, bit int32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

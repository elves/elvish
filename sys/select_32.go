// +build darwin 386,freebsd arm,freebsd 386,linux arm,linux netbsd openbsd

// The type of FdSet.Bits is different on different platforms.
// This file is for those where FdSet.Bits is []int32.

package sys

const NFDBits = 32

func index(fd int) (idx uint, bit int32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

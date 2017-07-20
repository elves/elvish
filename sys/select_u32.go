// +build netbsd openbsd arm,freebsd 386,freebsd

// The type of FdSet.Bits is different on different platforms.
// This file is for those where FdSet.Bits is []uint32.

package sys

const NFDBits = 32

func index(fd int) (idx uint, bit uint32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

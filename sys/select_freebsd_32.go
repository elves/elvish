// +build arm,freebsd 386,freebsd

// The type of FdSet.Bits is different on different platforms.
// This file is for FreeBSD where FdSet.Bits is actually X__fds_bits and it's []uint32.

package sys

const NFDBits = 32

func index(fd int) (idx uint, bit uint32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

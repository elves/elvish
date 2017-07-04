// +build amd64,freebsd arm64,freebsd

// The type of FdSet.Bits is different on different platforms.
// This file is for FreeBSD where FdSet.Bits is actually X__fds_bits and it's []uint64.

package sys

const NFDBits = 64

func index(fd int) (idx uint, bit uint64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

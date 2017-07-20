// +build linux
// +build amd64 arm64 mips64 mips64le ppc64 ppc64le s390x

// The type of FdSet.Bits is different on different platforms.
// This file is for those where FdSet.Bits is []int64.

package sys

const NFDBits = 64

func index(fd int) (idx uint, bit int64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

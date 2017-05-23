// +build amd64,dragonfly amd64,freebsd amd64,linux arm64,linux

// The type of FdSet.Bits is different on different platforms.
// This file is for those where FdSet.Bits is []int64.

package sys

const NFDBits = 64

func index(fd int) (idx uint, bit int64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

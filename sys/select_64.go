// +build amd64,dragonfly amd64,freebsd amd64,linux arm64,linux

// Differnt architectures lead to different bits of uint
// and different types of FdSet.Bits[].
// Hence, NFDBits and types of func index's return values
// have to be specified from case to case.

package sys

const NFDBits = 64

func index(fd int) (idx uint, bit int64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

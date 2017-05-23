// +build darwin 386,freebsd arm,freebsd 386,linux arm,linux netbsd openbsd

// Differnt architectures lead to different bits of uint
// and different types of FdSet.Bits[].
// Hence, NFDBits and types of func index's return values
// have to be specified from case to case.

package sys

const NFDBits = 32

func index(fd int) (idx uint, bit int32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

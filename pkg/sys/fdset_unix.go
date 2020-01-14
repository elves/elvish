// +build !windows,!plan9

package sys

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

var nFdBits = uint(8 * unsafe.Sizeof(unix.FdSet{}.Bits[0]))

type FdSet unix.FdSet

func (fs *FdSet) s() *unix.FdSet {
	return (*unix.FdSet)(fs)
}

func NewFdSet(fds ...int) *FdSet {
	fs := &FdSet{}
	fs.Set(fds...)
	return fs
}

func (fs *FdSet) Clear(fds ...int) {
	for _, fd := range fds {
		u := uint(fd)
		fs.Bits[u/nFdBits] &= ^(1 << (u % nFdBits))
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	u := uint(fd)
	return fs.Bits[u/nFdBits]&(1<<(u%nFdBits)) != 0
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		u := uint(fd)
		fs.Bits[u/nFdBits] |= 1 << (u % nFdBits)
	}
}

func (fs *FdSet) Zero() {
	*fs = FdSet{}
}

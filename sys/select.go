package sys

import (
	"syscall"
)

type FdSet syscall.FdSet

const NFDBits = 64

func index(fd int) (idx uint, bit int64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

func (fs *FdSet) s() *syscall.FdSet {
	return (*syscall.FdSet)(fs)
}

func NewFdSet(fds ...int) *FdSet {
	fs := &FdSet{}
	fs.Set(fds...)
	return fs
}

func (fs *FdSet) Clear(fds ...int) {
	for _, fd := range fds {
		idx, bit := index(fd)
		fs.Bits[idx] &= ^bit
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	idx, bit := index(fd)
	return fs.Bits[idx]&bit != 0
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		idx, bit := index(fd)
		fs.Bits[idx] |= bit
	}
}

func (fs *FdSet) Zero() {
	*fs = FdSet{}
}

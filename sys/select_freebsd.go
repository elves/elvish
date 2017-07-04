// +build freebsd

// For whatever reason, it's X__fds_bits instead of Bits on FreeBSD

package sys

import "syscall"

type FdSet syscall.FdSet

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
		fs.X__fds_bits[idx] &= ^bit
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	idx, bit := index(fd)
	return fs.X__fds_bits[idx]&bit != 0
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		idx, bit := index(fd)
		fs.X__fds_bits[idx] |= bit
	}
}

func (fs *FdSet) Zero() {
	*fs = FdSet{}
}

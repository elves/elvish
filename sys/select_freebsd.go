// +build freebsd

// For whatever reason, it's X__fds_bits instead of Bits on FreeBSD

package sys

import (
	"reflect"
	"syscall"
)

var nFdBits = (uint)(reflect.TypeOf(syscall.FdSet{}.X__fds_bits[0]).Size() * 8)

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
		u := uint(fd)
		fs.X__fds_bits[u/nFdBits] &= ^(1 << (u % nFdBits))
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	u := uint(fd)
	return fs.X__fds_bits[u/nFdBits]&(1<<(u%nFdBits)) != 0
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		u := uint(fd)
		fs.X__fds_bits[u/nFdBits] |= 1 << (u % nFdBits)
	}
}

func (fs *FdSet) Zero() {
	*fs = FdSet{}
}

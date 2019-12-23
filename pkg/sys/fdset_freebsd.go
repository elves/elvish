// For whatever reason, on FreeBSD the only field of FdSet is called
// X__fds_bits; on other Unices it is called Bits. This difference is irrelevant
// for C programs, as POSIX defines a set of macros for accessing FdSet, which
// hide the underlying difference. However since Elvish does not cgo and relies
// on the auto-generated struct definitions, it has to cope with the difference.

package sys

import (
	"reflect"

	"golang.org/x/sys/unix"
)

var nFdBits = (uint)(reflect.TypeOf(unix.FdSet{}.X__fds_bits[0]).Size() * 8)

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

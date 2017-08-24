// +build gccgo

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
		syscall.FDClr(fd, fs.s())
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	return syscall.FDIsSet(fd, fs.s())
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		syscall.FDSet(fd, fs.s())
	}
}

func (fs *FdSet) Zero() {
	syscall.FDZero(fs.s())
}

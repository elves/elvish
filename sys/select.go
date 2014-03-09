package sys

/*
#include <sys/select.h>

void fdclr(int fd, fd_set *set) {
	FD_CLR(fd, set);
}

int fdisset(int fd, fd_set *set) {
	return FD_ISSET(fd, set);
}

void fdset(int fd, fd_set *set) {
	FD_SET(fd, set);
}

void fdzero(fd_set *set) {
	FD_ZERO(set);
}
*/
import "C"

import (
	"syscall"
	"unsafe"
)

type FdSet syscall.FdSet

func (fs *FdSet) c() *C.fd_set {
	return (*C.fd_set)(unsafe.Pointer(fs))
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
		C.fdclr(C.int(fd), fs.c())
	}
}

func (fs *FdSet) IsSet(fd int) bool {
	return C.fdisset(C.int(fd), fs.c()) != 0
}

func (fs *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		C.fdset(C.int(fd), fs.c())
	}
}

func (fs *FdSet) Zero() {
	C.fdzero(fs.c())
}

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) (n int, err error) {
	return syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
}

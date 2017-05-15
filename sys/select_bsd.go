// +build darwin dragonfly freebsd netbsd openbsd

package sys

import "syscall"

const NFDBits = 32

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) (err error) {
	return syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
}

func index(fd int) (idx uint, bit int32) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

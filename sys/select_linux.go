// +build linux

package sys

import "syscall"

const NFDBits = 64

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) error {
	_, err := syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
	return err
}

func index(fd int) (idx uint, bit int64) {
	u := uint(fd)
	return u / NFDBits, 1 << (u % NFDBits)
}

// +build linux

package sys

import "syscall"

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) error {
	_, err := syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
	return err
}

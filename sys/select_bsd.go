// +build darwin dragonfly freebsd netbsd openbsd

package sys

import "syscall"

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) (err error) {
	return syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
}

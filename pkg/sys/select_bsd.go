// +build darwin dragonfly freebsd netbsd openbsd

package sys

import (
	"time"

	"golang.org/x/sys/unix"
)

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout time.Duration) error {
	var ptimeval *unix.Timeval
	if timeout >= 0 {
		timeval := unix.NsecToTimeval(int64(timeout))
		ptimeval = &timeval
	}
	return unix.Select(nfd, r.s(), w.s(), e.s(), ptimeval)
}

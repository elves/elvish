// +build linux

package sys

import "syscall"

func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *syscall.Timeval) error {
	// On ARM64, MIPS64 and MIPS64LE, syscall.Select is emulated in userland and
	// will dereference timeout. In that case, if the timeout argument is nil,
	// the call will panic. This is not POSIX-conformant behavior, but we work
	// around this by supplying a default value for timeout.
	_, err := syscall.Select(nfd, r.s(), w.s(), e.s(), timeout)
	return err
}

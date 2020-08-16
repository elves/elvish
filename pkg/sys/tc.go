// +build !windows,!plan9

package sys

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func Tcgetpgrp(fd int) (int, error) {
	var pid int
	errno := Ioctl(fd, unix.TIOCGPGRP, uintptr(unsafe.Pointer(&pid)))
	if errno == nil {
		return pid, nil
	}
	return -1, errno
}

func Tcsetpgrp(fd int, pid int) error {
	return Ioctl(fd, unix.TIOCSPGRP, uintptr(unsafe.Pointer(&pid)))
}

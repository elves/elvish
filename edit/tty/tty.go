// Package tty wraps tty ioctls.
package tty

import (
	"os"
	"syscall"
)

func Ioctl(fd int, req int, arg uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL, uintptr(fd), uintptr(req), arg)
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	} else {
		return nil
	}
}

func FlushInput(fd int) error {
	return Ioctl(fd, TCFLSH, TCIFLUSH)
}

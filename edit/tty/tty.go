// Package tty wraps tty ioctls.
package tty

/*
#include <termios.h>
*/
import "C"
import (
	"os"
	"syscall"
)

func Ioctl(fd int, req int, arg uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL, uintptr(fd), uintptr(req), arg)
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	}
	return nil
}

func FlushInput(fd int) error {
	_, err := C.tcflush((C.int)(fd), TCIFLUSH)
	return err
}

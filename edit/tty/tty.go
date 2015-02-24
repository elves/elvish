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

// Ioctl wraps the ioctl syscall.
func Ioctl(fd int, req int, arg uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL, uintptr(fd), uintptr(req), arg)
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	}
	return nil
}

// FlushInput discards data written to a file descriptor but not read.
func FlushInput(fd int) error {
	_, err := C.tcflush((C.int)(fd), syscall.TCIFLUSH)
	return err
}

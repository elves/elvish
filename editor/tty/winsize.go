// +build ignore

package tty

/*
#include <termios.h>
#include <sys/ioctl.h>
*/
import "C"
import (
	"unsafe"
)

type Winsize C.struct_winsize

func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, C.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return ws
}

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

const (
	TCFLSH     = C.TCFLSH
	TCIFLUSH   = C.TCIFLUSH
	TIOCGWINSZ = C.TIOCGWINSZ
)

type Winsize C.struct_winsize

func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return ws
}

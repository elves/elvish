// +build ignore

package sys

/*
#include <termios.h>
#include <sys/ioctl.h>
*/
import "C"
import (
	"syscall"
	"unsafe"
)

// Winsize wraps the C winsize struct and represents the size of a terminal.
type Winsize C.struct_winsize

// GetWinsize queries the size of the terminal referenced by the given file
// descriptor.
func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return ws
}

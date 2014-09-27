package tty

/*
#include <termios.h>
*/
import "C"
import (
	"syscall"
	"unsafe"
)

type Termios syscall.Termios

func NewTermiosFromFd(fd int) (*Termios, error) {
	term := new(Termios)
	err := term.FromFd(fd)
	if err != nil {
		return nil, err
	}
	return term, nil
}

func (term *Termios) FromFd(fd int) error {
	_, err := C.tcgetattr((C.int)(fd),
		(*C.struct_termios)(unsafe.Pointer(term)))
	return err
}

func (term *Termios) ApplyToFd(fd int) error {
	_, err := C.tcsetattr((C.int)(fd), 0,
		(*C.struct_termios)(unsafe.Pointer(term)))
	return err
}

func (term *Termios) Copy() *Termios {
	v := *term
	return &v
}

func (term *Termios) SetTime(v uint8) {
	term.Cc[syscall.VTIME] = v
}

func (term *Termios) SetMin(v uint8) {
	term.Cc[syscall.VMIN] = v
}

func setFlag(flag *uint32, mask uint32, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

func (term *Termios) SetIcanon(v bool) {
	setFlag((*uint32)(unsafe.Pointer(&term.Lflag)), syscall.ICANON, v)
}

func (term *Termios) SetEcho(v bool) {
	setFlag((*uint32)(unsafe.Pointer(&term.Lflag)), syscall.ECHO, v)
}

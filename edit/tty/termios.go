package tty

/*
#include <termios.h>
*/
import "C"

type Termios C.struct_termios

func NewTermiosFromFd(fd int) (*Termios, error) {
	term := new(Termios)
	err := term.FromFd(fd)
	if err != nil {
		return nil, err
	}
	return term, nil
}

func (term *Termios) c() *C.struct_termios {
	return (*C.struct_termios)(term)
}

func (term *Termios) FromFd(fd int) error {
	_, err := C.tcgetattr((C.int)(fd), term.c())
	return err
}

func (term *Termios) ApplyToFd(fd int) error {
	_, err := C.tcsetattr((C.int)(fd), 0, term.c())
	return err
}

func (term *Termios) Copy() *Termios {
	v := *term
	return &v
}

func (term *Termios) SetTime(v uint8) {
	term.c_cc[C.VTIME] = C.cc_t(v)
}

func (term *Termios) SetMin(v uint8) {
	term.c_cc[C.VMIN] = C.cc_t(v)
}

func setFlag(flag *C.tcflag_t, mask C.tcflag_t, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

func (term *Termios) SetIcanon(v bool) {
	setFlag(&term.c_lflag, C.ICANON, v)
}

func (term *Termios) SetEcho(v bool) {
	setFlag(&term.c_lflag, C.ECHO, v)
}

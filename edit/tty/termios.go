package tty

/*
#include <termios.h>
*/
import "C"

// Termios represents terminal attributes.
type Termios C.struct_termios

// NewTermiosFromFd extracts the terminal attribute of the given file
// descriptor.
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

// FromFd fills term with the terminal attribute of the given file descriptor.
func (term *Termios) FromFd(fd int) error {
	_, err := C.tcgetattr((C.int)(fd), term.c())
	return err
}

// ApplyToFd applies term to the given file descriptor.
func (term *Termios) ApplyToFd(fd int) error {
	_, err := C.tcsetattr((C.int)(fd), 0, term.c())
	return err
}

// Copy returns a copy of term.
func (term *Termios) Copy() *Termios {
	v := *term
	return &v
}

// SetVTime sets the timeout in deciseconds for noncanonical read.
func (term *Termios) SetVTime(v uint8) {
	term.c_cc[C.VTIME] = C.cc_t(v)
}

// SetVMin sets the minimal number of characters for noncanonical read.
func (term *Termios) SetVMin(v uint8) {
	term.c_cc[C.VMIN] = C.cc_t(v)
}

func setFlag(flag *C.tcflag_t, mask C.tcflag_t, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

// SetICanon sets the canonical flag.
func (term *Termios) SetICanon(v bool) {
	setFlag(&term.c_lflag, C.ICANON, v)
}

// SetEcho sets the echo flag.
func (term *Termios) SetEcho(v bool) {
	setFlag(&term.c_lflag, C.ECHO, v)
}

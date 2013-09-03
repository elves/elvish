package tty

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
	return Ioctl(fd, syscall.TCGETS, uintptr(unsafe.Pointer(term)))
}

func (term *Termios) ApplyToFd(fd int) error {
	return Ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(term)))
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
	setFlag(&term.Lflag, syscall.ICANON, v)
}

func (term *Termios) SetEcho(v bool) {
	setFlag(&term.Lflag, syscall.ECHO, v)
}

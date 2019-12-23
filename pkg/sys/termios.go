// +build !windows,!plan9

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// Termios represents terminal attributes.
type Termios unix.Termios

// NewTermiosFromFd extracts the terminal attribute of the given file
// descriptor.
func NewTermiosFromFd(fd int) (*Termios, error) {
	var term Termios
	if err := Ioctl(fd, getAttrIOCTL, uintptr(unsafe.Pointer(&term))); err != nil {
		return nil, err
	}
	return &term, nil
}

// ApplyToFd applies term to the given file descriptor.
func (term *Termios) ApplyToFd(fd int) error {
	return Ioctl(fd, setAttrNowIOCTL, uintptr(unsafe.Pointer(term)))
}

// Copy returns a copy of term.
func (term *Termios) Copy() *Termios {
	v := *term
	return &v
}

// SetVTime sets the timeout in deciseconds for noncanonical read.
func (term *Termios) SetVTime(v uint8) {
	term.Cc[unix.VTIME] = v
}

// SetVMin sets the minimal number of characters for noncanonical read.
func (term *Termios) SetVMin(v uint8) {
	term.Cc[unix.VMIN] = v
}

// SetICanon sets the canonical flag.
func (term *Termios) SetICanon(v bool) {
	setFlag(&term.Lflag, unix.ICANON, v)
}

// SetEcho sets the echo flag.
func (term *Termios) SetEcho(v bool) {
	setFlag(&term.Lflag, unix.ECHO, v)
}

// SetICRNL sets the CRNL iflag bit
func (term *Termios) SetICRNL(v bool) {
	setFlag(&term.Iflag, unix.ICRNL, v)
}

// FlushInput discards data written to a file descriptor but not read.
func FlushInput(fd int) error {
	return Ioctl(fd, flushIOCTL, uintptr(unix.TCIFLUSH))
}

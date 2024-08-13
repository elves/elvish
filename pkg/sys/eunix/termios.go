//go:build unix

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package eunix

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// Termios represents terminal attributes.
type Termios unix.Termios

// TermiosForFd returns a pointer to a Termios structure if the file
// descriptor is open on a terminal device.
func TermiosForFd(fd int) (*Termios, error) {
	term, err := unix.IoctlGetTermios(fd, getAttrIOCTL)
	return (*Termios)(term), err
}

// ApplyToFd applies term to the given file descriptor.
func (term *Termios) ApplyToFd(fd int) error {
	return unix.IoctlSetTermios(fd, setAttrNowIOCTL, (*unix.Termios)(unsafe.Pointer(term)))
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

// SetIExten sets the iexten flag.
func (term *Termios) SetIExten(v bool) {
	setFlag(&term.Lflag, unix.IEXTEN, v)
}

// SetEcho sets the echo flag.
func (term *Termios) SetEcho(v bool) {
	setFlag(&term.Lflag, unix.ECHO, v)
}

// SetICRNL sets the CRNL iflag bit.
func (term *Termios) SetICRNL(v bool) {
	setFlag(&term.Iflag, unix.ICRNL, v)
}

// SetIXON sets the IXON iflag bit.
func (term *Termios) SetIXON(v bool) {
	setFlag(&term.Iflag, unix.IXON, v)
}

func setFlag(flag *termiosFlag, mask termiosFlag, v bool) {
	if v {
		*flag |= mask
	} else {
		*flag &= ^mask
	}
}

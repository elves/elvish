//go:build darwin || dragonfly || freebsd || netbsd || openbsd

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package eunix

import "golang.org/x/sys/unix"

const (
	getAttrIOCTL      = unix.TIOCGETA
	setAttrNowIOCTL   = unix.TIOCSETA
	setAttrDrainIOCTL = unix.TIOCSETAW
	setAttrFlushIOCTL = unix.TIOCSETAF
	flushIOCTL        = unix.TIOCFLUSH
)

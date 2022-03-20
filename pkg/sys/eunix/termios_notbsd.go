//go:build linux || solaris

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package eunix

import "golang.org/x/sys/unix"

const (
	getAttrIOCTL      = unix.TCGETS
	setAttrNowIOCTL   = unix.TCSETS
	setAttrDrainIOCTL = unix.TCSETSW
	setAttrFlushIOCTL = unix.TCSETSF
	flushIOCTL        = unix.TCFLSH
)

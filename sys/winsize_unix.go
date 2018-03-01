// +build !windows,!plan9

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SIGWINCH is the Window size change signal.
const SIGWINCH = unix.SIGWINCH

// winSize mirrors struct winsize in the C header.
// The following declaration matches struct winsize in the headers of
// Linux and FreeBSD.
type winSize struct {
	row    uint16
	col    uint16
	Xpixel uint16
	Ypixel uint16
}

// GetWinsize queries the size of the terminal referenced by the given file.
func GetWinsize(file *os.File) (row, col int) {
	fd := int(file.Fd())
	ws := winSize{}
	if err := Ioctl(fd, unix.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws))); err != nil {
		fmt.Printf("error in winSize: %v", err)
		return -1, -1
	}

	// Pick up a reasonable value for row and col
	// if they equal zero in special case,
	// e.g. serial console
	if ws.col == 0 {
		ws.col = 80
	}
	if ws.row == 0 {
		ws.row = 24
	}

	return int(ws.row), int(ws.col)
}

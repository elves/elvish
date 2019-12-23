// +build !windows,!plan9

// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// SIGWINCH is the Window size change signal.
const SIGWINCH = unix.SIGWINCH

// GetWinsize queries the size of the terminal referenced by the given file.
func GetWinsize(file *os.File) (row, col int) {
	fd := int(file.Fd())
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		fmt.Printf("error in winSize: %v", err)
		return -1, -1
	}

	// Pick up a reasonable value for row and col
	// if they equal zero in special case,
	// e.g. serial console
	if ws.Col == 0 {
		ws.Col = 80
	}
	if ws.Row == 0 {
		ws.Row = 24
	}

	return int(ws.Row), int(ws.Col)
}

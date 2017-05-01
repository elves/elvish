// Copyright 2015 go-termios Author. All Rights Reserved.
// https://github.com/go-termios/termios
// Author: John Lenton <chipaca@github.com>

package sys

import (
	"fmt"
	"golang.org/x/sys/unix"
	"unsafe"
)

// GetWinsize queries the size of the terminal referenced by the given file
// descriptor.
//
type winSize struct {
	row    uint16
	col    uint16
	Xpixel uint16
	Ypixel uint16
}

func GetWinsize(fd int) (row, col int) {
	ws := winSize{}
	if err := ioctl(uintptr(fd), unix.TIOCGWINSZ, unsafe.Pointer(&ws)); err != nil {
		fmt.Printf("error in winSize: %v", err)
		return -1, -1
	}
	return int(ws.row), int(ws.col)
}

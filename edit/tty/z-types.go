// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs edit/tty/types.go

package tty

import (
	"unsafe"
)

const (
	TCFLSH		= 0x540b
	TCIFLUSH	= 0x0
	TIOCGWINSZ	= 0x5413
)

type Winsize struct {
	Row	uint16
	Col	uint16
	Xpixel	uint16
	Ypixel	uint16
}

func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return ws
}

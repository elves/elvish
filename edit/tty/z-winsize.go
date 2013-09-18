// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs edit/tty/winsize.go

package tty

import (
	"unsafe"
)

type Winsize struct {
	Row	uint16
	Col	uint16
	Xpixel	uint16
	Ypixel	uint16
}

func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, 0x5413, uintptr(unsafe.Pointer(&ws)))
	return ws
}

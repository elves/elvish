// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs winsize.go

package sys

import (
	"syscall"
	"unsafe"
)

type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func GetWinsize(fd int) Winsize {
	var ws Winsize
	Ioctl(fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return ws
}

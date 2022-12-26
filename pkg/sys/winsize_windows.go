package sys

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// Windows doesn't have SIGCH, so use an impossible value.
const sigWINCH = syscall.Signal(-1)

func winSize(file *os.File) (row, col int) {
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(file.Fd()), &info)
	if err != nil {
		return -1, -1
	}
	window := info.Window
	return int(window.Bottom - window.Top), int(window.Right - window.Left)
}

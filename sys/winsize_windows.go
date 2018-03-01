package sys

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// SIGWINCH is the Window size change signal. On Windows this signal does not
// exist, so we use -1, an impossible value for signals.
// NOTE: This has to use the syscall package before the x/sys/* packages also
// use syscall.Signal to define signals.
const SIGWINCH = syscall.Signal(-1)

// GetWinsize queries the size of the terminal referenced by the given file.
func GetWinsize(file *os.File) (row, col int) {
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(file.Fd()), &info)
	if err != nil {
		fmt.Printf("error in winSize: %v", err)
		return -1, -1
	}
	window := info.Window
	return int(window.Bottom - window.Top), int(window.Right - window.Left)
}

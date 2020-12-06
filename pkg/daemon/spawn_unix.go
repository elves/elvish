// +build !windows,!plan9,!js

package daemon

import (
	"os"
	"syscall"
)

func procAttrForSpawn(stdout *os.File) *os.ProcAttr {
	// The daemon should not inherit a reference to the stdin file descriptor
	// of the original shell. The daemon shouldn't read anything from stdin so
	// use the null device.
	devnull, _ := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	return &os.ProcAttr{
		Dir:   "/",
		Env:   []string{},
		Files: []*os.File{devnull, stdout, stdout},
		Sys: &syscall.SysProcAttr{
			Setsid: true, // detach from current terminal
		},
	}
}

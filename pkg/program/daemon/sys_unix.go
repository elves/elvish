// +build !windows
// +build !plan9

package daemon

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func setUmask() {
	unix.Umask(0077)
}

func procAttrForSpawn() *os.ProcAttr {
	return &os.ProcAttr{
		Dir:   "/",        // cd to /
		Env:   []string{}, // empty environment
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys: &syscall.SysProcAttr{
			Setsid: true, // detach from current terminal
		},
	}
}

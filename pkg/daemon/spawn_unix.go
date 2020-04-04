// +build !windows,!plan9,!js

package daemon

import (
	"os"
	"syscall"
)

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

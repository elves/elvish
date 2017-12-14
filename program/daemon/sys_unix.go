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

func proAttrForFirstFork() *os.ProcAttr {
	return &os.ProcAttr{
		Dir: "/",        // cd to /
		Env: []string{}, // empty environment
		Sys: &syscall.SysProcAttr{Setsid: true},
	}
}

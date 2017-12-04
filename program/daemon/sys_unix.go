// +build !windows
// +build !plan9

package daemon

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func setUmask() {
	unix.Umask(0077)
}

func sysProAttrForFirstFork() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

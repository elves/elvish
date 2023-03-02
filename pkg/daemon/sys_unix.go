//go:build unix

package daemon

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

var errConnRefused = syscall.ECONNREFUSED

// Make sure that files created by the daemon is not accessible to other users.
func setUmaskForDaemon() { unix.Umask(0077) }

func procAttrForSpawn(files []*os.File) *os.ProcAttr {
	return &os.ProcAttr{
		Dir:   "/",
		Env:   []string{},
		Files: files,
		Sys: &syscall.SysProcAttr{
			Setsid: true, // detach from current terminal
		},
	}
}

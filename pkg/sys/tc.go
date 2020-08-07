// +build !windows,!plan9

package sys

import (
	"golang.org/x/sys/unix"
)

func Tcsetpgrp(fd int, pid int) error {
	return unix.IoctlSetInt(fd, unix.TIOCSPGRP, pid)
}

// +build !windows,!plan9

package sys

import (
	"syscall"
)

func Fcntl(fd int, cmd int, arg int) (val int, err error) {
	r, _, e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd),
		uintptr(arg))
	val = int(r)
	if e != 0 {
		err = e
	}
	return
}

func GetNonblock(fd int) (bool, error) {
	r, err := Fcntl(fd, syscall.F_GETFL, 0)
	return r&syscall.O_NONBLOCK != 0, err
}

func SetNonblock(fd int, nonblock bool) error {
	r, err := Fcntl(fd, syscall.F_GETFL, 0)
	if err != nil {
		return err
	}
	if nonblock {
		r |= syscall.O_NONBLOCK
	} else {
		r &^= syscall.O_NONBLOCK
	}
	_, err = Fcntl(fd, syscall.F_SETFL, r)
	return err
}

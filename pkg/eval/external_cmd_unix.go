//go:build unix

package eval

import "syscall"

func isSIGPIPE(s syscall.Signal) bool {
	return s == syscall.SIGPIPE
}

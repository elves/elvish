package eval

import "syscall"

func isSIGPIPE(s syscall.Signal) bool {
	// Windows doesn't have SIGPIPE.
	return false
}

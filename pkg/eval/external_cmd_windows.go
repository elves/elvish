package eval

import "syscall"

func isSIGPIPE(_ syscall.Signal) bool {
	// Windows doesn't have SIGPIPE.
	return false
}

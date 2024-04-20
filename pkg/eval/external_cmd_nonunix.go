//go:build !unix

package eval

import "syscall"

func isSIGPIPE(s syscall.Signal) bool {
	// Windows and WASM don't have SIGPIPE.
	return false
}

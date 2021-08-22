//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package eval

import "syscall"

func isSIGPIPE(s syscall.Signal) bool {
	return s == syscall.SIGPIPE
}

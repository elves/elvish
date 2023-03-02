//go:build unix

package eval

import "syscall"

var epipe = syscall.EPIPE

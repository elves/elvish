//go:build !windows

package eval

import "syscall"

var epipe = syscall.EPIPE

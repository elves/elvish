//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package eval

import "syscall"

var epipe = syscall.EPIPE

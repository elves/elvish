//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package testutil

import "golang.org/x/sys/unix"

var umask = unix.Umask

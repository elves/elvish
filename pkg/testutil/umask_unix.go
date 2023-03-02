//go:build unix

package testutil

import "golang.org/x/sys/unix"

var umask = unix.Umask

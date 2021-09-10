//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package lscolors

import (
	"testing"

	"golang.org/x/sys/unix"
)

func createNamedPipe(fname string) error {
	return unix.Mkfifo(fname, 0600)
}

func setUmask(t *testing.T, m int) {
	save := unix.Umask(m)
	t.Cleanup(func() { unix.Umask(save) })
}

//go:build unix

package lscolors

import (
	"golang.org/x/sys/unix"
)

func createNamedPipe(fname string) error {
	return unix.Mkfifo(fname, 0600)
}

//go:build unix

package os_test

import (
	"testing"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/must"
)

func mkFifoOrSkip(name string) { must.OK(unix.Mkfifo(name, 0o600)) }

func createWindowsSpecialFileOrSkip(t *testing.T) { t.Skip("not on Windows") }

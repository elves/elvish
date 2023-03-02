//go:build unix

package shell

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"src.elv.sh/pkg/sys"
)

func handleSignal(sig os.Signal, stderr io.Writer) {
	switch sig {
	case syscall.SIGHUP:
		syscall.Kill(0, syscall.SIGHUP)
		os.Exit(0)
	case syscall.SIGUSR1:
		fmt.Fprint(stderr, sys.DumpStack())
	}
}

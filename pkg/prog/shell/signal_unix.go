// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/pkg/sys"
)

func handleSignal(sig os.Signal, stderr *os.File) {
	switch sig {
	case syscall.SIGHUP:
		syscall.Kill(0, syscall.SIGHUP)
		os.Exit(0)
	case syscall.SIGUSR1:
		fmt.Fprint(stderr, sys.DumpStack())
	}
}

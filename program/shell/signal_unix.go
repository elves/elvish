// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/sys"
)

func handleSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGHUP:
		syscall.Kill(0, syscall.SIGHUP)
		os.Exit(0)
	case syscall.SIGUSR1:
		fmt.Print(sys.DumpStack())
	}
}

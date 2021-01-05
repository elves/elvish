// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/sys"
)

func ignoreSignal(sig os.Signal) bool {
	// SIGURG isn't interesting since it is used internally by the Go runtime on UNIX and occurs
	// with great frequency.
	return sig.(syscall.Signal) == syscall.SIGURG
}

func signalName(sig os.Signal) string {
	return unix.SignalName(sig.(syscall.Signal))
}

func handleSignal(sig os.Signal, stderr *os.File) {
	switch sig {
	case syscall.SIGHUP:
		syscall.Kill(0, syscall.SIGHUP)
		os.Exit(0)
	case syscall.SIGUSR1:
		fmt.Fprint(stderr, sys.DumpStack())
	}
}

package shell

import (
	"io"
	"os"
	"syscall"
)

func handleSignal(sig os.Signal, stderr io.Writer) {
	switch sig {
	// See https://pkg.go.dev/os/signal#hdr-Windows for the semantics of SIGTERM
	// on Windows.
	case syscall.SIGTERM:
		os.Exit(0)
	}
}

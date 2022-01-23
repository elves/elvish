package shell

import (
	"io"
	"os"
	"syscall"
)

func handleSignal(sig os.Signal, stderr io.Writer) {
	switch sig {
	case syscall.SIGTERM:
		os.Exit(0)
	}
}

package shell

import (
	"os"
)

func ignoreSignal(sig os.Signal) bool {
	return false
}

func signalName(sig os.Signal) string {
	return "SIG???"
}

func handleSignal(os.Signal, *os.File) {
}

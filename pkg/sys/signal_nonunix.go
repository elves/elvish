//go:build windows || plan9 || js

package sys

import (
	"os"
	"os/signal"
)

func notifySignals() chan os.Signal {
	// This catches every signal regardless of whether it is ignored.
	sigCh := make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(sigCh)
	return sigCh
}

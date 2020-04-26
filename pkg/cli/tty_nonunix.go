// +build windows plan9 js

package cli

import (
	"os"
	"os/signal"
)

func (t *aTTY) NotifySignals() <-chan os.Signal {
	// This implicitly catches every signal regardless of whether it is
	// currently ignored.
	t.sigCh = make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(t.sigCh)
	return t.sigCh
}

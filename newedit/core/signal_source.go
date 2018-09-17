package core

import (
	"os"
	"os/signal"
)

// SignalSource is used by the editor to obtain signals.
type SignalSource interface {
	// NotifySignals cause the SignalSource to start relaying signals, and
	// returns a channel on which signals are delivered.
	NotifySignals() <-chan os.Signal
	// StopSignals stops listening to signals. After the function returns, the
	// channel returned by NotifySignals may not deliver any value any more.
	StopSignals()
}

const sigsChanBufferSize = 256

type signalSource struct {
	sigs []os.Signal
	ch   chan os.Signal
}

// NewSignalSource creates a SignalSource that delivers the given signals using
// the os/signal package.
func NewSignalSource(sigs ...os.Signal) SignalSource {
	return &signalSource{sigs, nil}
}

func (sigs *signalSource) NotifySignals() <-chan os.Signal {
	sigs.ch = make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(sigs.ch, sigs.sigs...)
	return sigs.ch
}

func (sigs *signalSource) StopSignals() {
	signal.Stop(sigs.ch)
	close(sigs.ch)
	sigs.ch = nil
}

package clicore

import "os"

const maxFakeSignals = 1024

// FakeSignalSource is an implementation of SignalSource that is useful in
// tests.
type FakeSignalSource struct {
	// A channel on which fake signals can be injected.
	Ch chan os.Signal
}

// NewFakeSignalSource creates a new FakeSignalSource.
func NewFakeSignalSource() *FakeSignalSource {
	return &FakeSignalSource{make(chan os.Signal, maxFakeSignals)}
}

// NotifySignals returns sigs.Ch.
func (sigs *FakeSignalSource) NotifySignals() <-chan os.Signal {
	return sigs.Ch
}

// StopSignals closes sig.Ch and set it to nil.
func (sigs *FakeSignalSource) StopSignals() {
	close(sigs.Ch)
	sigs.Ch = nil
}

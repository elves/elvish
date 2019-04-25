package clicore

import "os"

const maxFakeSignals = 1024

// An implementation of SignalSource that is useful in tests.
type fakeSignalSource struct {
	// A channel on which fake signals can be injected.
	Ch chan os.Signal
}

// Creates a new FakeSignalSource.
func newFakeSignalSource() *fakeSignalSource {
	return &fakeSignalSource{make(chan os.Signal, maxFakeSignals)}
}

// NotifySignals returns sigs.Ch.
func (sigs *fakeSignalSource) NotifySignals() <-chan os.Signal {
	return sigs.Ch
}

// StopSignals closes sig.Ch and set it to nil.
func (sigs *fakeSignalSource) StopSignals() {
	close(sigs.Ch)
	sigs.Ch = nil
}

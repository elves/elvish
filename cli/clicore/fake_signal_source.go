package clicore

import "os"

const maxFakeSignals = 1024

// SignalSourceCtrl is an interface for controlling the fake signal source.
type SignalSourceCtrl interface {
	// Inject injects a signal.
	Inject(sig os.Signal)
}

// An implementation of SignalSource that is useful in tests.
type fakeSignalSource chan os.Signal

// NewFakeSignalSource creates a new FakeSignalSource.
func NewFakeSignalSource() (SignalSource, SignalSourceCtrl) {
	ch := make(chan os.Signal, maxFakeSignals)
	return fakeSignalSource(ch), signalSourceCtrl(ch)
}

func (sigs fakeSignalSource) NotifySignals() <-chan os.Signal { return sigs }

func (sigs fakeSignalSource) StopSignals() { close(sigs) }

type signalSourceCtrl chan<- os.Signal

func (sigs signalSourceCtrl) Inject(sig os.Signal) { sigs <- sig }

package core

import "os"

const maxFakeSignals = 1024

type fakeSignalSource struct {
	ch chan os.Signal
}

func newFakeSignalSource() *fakeSignalSource {
	return &fakeSignalSource{make(chan os.Signal, maxFakeSignals)}
}

func (sigs *fakeSignalSource) NotifySignals() <-chan os.Signal {
	return sigs.ch
}

func (sigs *fakeSignalSource) StopSignals() {
	close(sigs.ch)
	sigs.ch = nil
}

package core

import (
	"os"
	"os/signal"
)

type SignalSource interface {
	NotifySignals() <-chan os.Signal
	StopSignals()
}

const sigsChanBufferSize = 256

type signalSource struct {
	sigs []os.Signal
	ch   chan os.Signal
}

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

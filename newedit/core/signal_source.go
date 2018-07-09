package core

import (
	"os"
)

type SignalSource interface {
	NotifySignals() <-chan os.Signal
	StopSignals()
}

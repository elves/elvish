// +build !windows,!plan9

package core

import (
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const (
	signalTimeout    = 10 * time.Millisecond
	signalNegTimeout = 20 * time.Millisecond
)

func TestSignalSource(t *testing.T) {
	sigs := NewSignalSource(unix.SIGUSR1)
	sigch := sigs.NotifySignals()

	err := unix.Kill(unix.Getpid(), unix.SIGUSR1)
	if err != nil {
		t.Skip("cannot send SIGUSR1 to myself:", err)
	}

	select {
	case sig := <-sigch:
		if sig != unix.SIGUSR1 {
			t.Errorf("Received signal %v, want SIGUSR1", sig)
		}
	case <-time.After(signalTimeout):
		t.Errorf("Timeout waiting for signal relay")
	}

	sigs.StopSignals()

	err = unix.Kill(unix.Getpid(), unix.SIGUSR1)
	if err != nil {
		t.Skip("cannot send SIGUSR1 to myself:", err)
	}

	select {
	case sig := <-sigch:
		t.Errorf("Still received signal %v after StopSignals", sig)
	case <-time.After(signalNegTimeout):
	}
}

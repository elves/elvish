// +build !windows,!plan9

package core

import (
	"os"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSignalSource(t *testing.T) {
	sigs := NewSignalSource(unix.SIGUSR1)
	sigch := sigs.NotifySignals()

	collectedCh := make(chan []os.Signal, 1)
	go func() {
		var collected []os.Signal
		for sig := range sigch {
			collected = append(collected, sig)
		}
		collectedCh <- collected
	}()

	err := unix.Kill(unix.Getpid(), unix.SIGUSR1)
	if err != nil {
		t.Skip("cannot send SIGUSR1 to myself:", err)
	}

	sigs.StopSignals()

	err = unix.Kill(unix.Getpid(), unix.SIGUSR2)
	if err != nil {
		t.Skip("cannot send SIGUSR2 to myself:", err)
	}

	collected := <-collectedCh
	wantCollected := []os.Signal{unix.SIGUSR1}
	if !reflect.DeepEqual(collected, wantCollected) {
		t.Errorf("collected %v, want %v", collected, wantCollected)
	}
}

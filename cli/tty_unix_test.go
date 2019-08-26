// +build !windows,!plan9

package cli

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestTTYSignal(t *testing.T) {
	tty := NewStdTTY()
	sigch := tty.NotifySignals()

	err := unix.Kill(unix.Getpid(), unix.SIGUSR1)
	if err != nil {
		t.Skip("cannot send SIGUSR1 to myself:", err)
	}

	if sig := <-sigch; sig != unix.SIGUSR1 {
		t.Errorf("Got signal %v, want SIGUSR1", sig)
	}

	tty.StopSignals()

	err = unix.Kill(unix.Getpid(), unix.SIGUSR2)
	if err != nil {
		t.Skip("cannot send SIGUSR2 to myself:", err)
	}

	if sig := <-sigch; sig != nil {
		t.Errorf("Got signal %v, want nil", sig)
	}
}

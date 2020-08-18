// +build !windows,!plan9,!js

package eval

import (
	"os"
	"testing"
	"time"
)

func interruptedTimeAfterMock(fm *Frame, d time.Duration) <-chan time.Time {
	if d == 10*time.Millisecond {
		// Special-case intended to verity that a sleep can be interrupted.
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)
		return time.After(1 * time.Second)
	}
	panic("unreachable")
}

func TestInterruptedSleep(t *testing.T) {
	timeAfter = interruptedTimeAfterMock
	Test(t,
		// Special-case that should result in the sleep being interrupted. See
		// timeAfterMock above.
		That(`sleep 10ms`).Throws(ErrInterrupted, "sleep 10ms"),
	)
}

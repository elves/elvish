// +build !windows,!plan9,!js

package eval_test

import (
	"os"
	"testing"
	"time"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/testutil"

	. "src.elv.sh/pkg/eval/evaltest"
)

func interruptedTimeAfterMock(fm *Frame, d time.Duration) <-chan time.Time {
	if d == time.Second {
		// Special-case intended to verity that a sleep can be interrupted.
		go func() {
			// Wait a little bit to ensure that the control flow in the "sleep"
			// function is in the select block when the interrupt is sent.
			time.Sleep(testutil.ScaledMs(1))
			p, _ := os.FindProcess(os.Getpid())
			p.Signal(os.Interrupt)
		}()
		return time.After(1 * time.Second)
	}
	panic("unreachable")
}

func TestInterruptedSleep(t *testing.T) {
	TimeAfter = interruptedTimeAfterMock
	Test(t,
		// Special-case that should result in the sleep being interrupted. See
		// timeAfterMock above.
		That(`sleep 1s`).Throws(ErrInterrupted, "sleep 1s"),
	)
}

//go:build unix

package eval_test

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/testutil"

	. "src.elv.sh/pkg/eval/evaltest"
)

func TestSleep_Interrupt(t *testing.T) {
	testutil.Set(t, TimeAfter,
		func(fm *Frame, d time.Duration) <-chan time.Time {
			go unix.Kill(os.Getpid(), unix.SIGINT)
			return time.After(d)
		})

	Test(t,
		That(`sleep 1s`).Throws(ErrInterrupted, "sleep 1s"),
	)
}

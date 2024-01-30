//go:build unix

package eval_test

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/testutil"
)

func injectTimeAfterWithSIGINTOrSkip(t *testing.T) {
	testutil.Set(t, eval.TimeAfter,
		func(_ *eval.Frame, d time.Duration) <-chan time.Time {
			go unix.Kill(os.Getpid(), unix.SIGINT)
			return time.After(d)
		})
}

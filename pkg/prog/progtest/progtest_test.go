package progtest

import (
	"testing"
)

// Verify we don't deadlock if more output is written to stdout than can be
// buffered by a pipe.
func TestOutputCaptureDoesNotDeadlock(t *testing.T) {
	t.Helper()
	f := Setup()
	defer f.Cleanup()

	// We need enough data to verify whether we're likely to deadlock due to
	// filling the pipe before the test completes. Pipes typically buffer 8 to
	// 128 KiB.
	bytes := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	for i := 0; i < 128*1024/len(bytes); i++ {
		f.pipes[1].w.Write(bytes[:])
	}
	f.pipes[1].w.WriteString("hello\n")
	f.TestOutSnippet(t, 1, "hello")
	f.TestOut(t, 2, "")
}

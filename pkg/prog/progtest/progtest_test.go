package progtest

import (
	"os"
	"testing"

	"src.elv.sh/pkg/prog"
)

// Verify we don't deadlock if more output is written to stdout than can be
// buffered by a pipe.
func TestOutputCaptureDoesNotDeadlock(t *testing.T) {
	Test(t, noisyProgram{},
		ThatElvish().WritesStdoutContaining("hello"),
	)
}

type noisyProgram struct{}

func (noisyProgram) RegisterFlags(f *prog.FlagSet) {}

func (noisyProgram) Run(fds [3]*os.File, args []string) error {
	// We need enough data to verify whether we're likely to deadlock due to
	// filling the pipe before the test completes. Pipes typically buffer 8 to
	// 128 KiB.
	bytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	for i := 0; i < 128*1024/len(bytes); i++ {
		fds[1].Write(bytes)
	}
	fds[1].WriteString("hello")
	return nil
}

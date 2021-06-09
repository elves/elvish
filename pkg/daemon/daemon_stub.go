// +build elv_daemon_stub

// This, and the other *_stub.go files, exists to make it possible to build an elvish binary with no
// support for an interactive database daemon. Thus producing a smaller program with less overhead.
// Which is valuable in "busybox" (e.g., u-root) type environments.

package daemon

import (
	"os"

	"src.elv.sh/pkg/prog"
)

const Version = 0

// Program is the daemon subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) ShouldRun(f *prog.Flags) bool { return false }

func (program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	return nil
}

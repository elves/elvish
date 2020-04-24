package prog

import (
	"os"

	"github.com/elves/elvish/pkg/daemon"
)

type DaemonProgram struct{}

func (DaemonProgram) ShouldRun(f *Flags) bool { return f.Daemon }

func (DaemonProgram) Run(fds [3]*os.File, f *Flags, args []string) error {
	if len(args) > 0 {
		return BadUsage("arguments are not allowed with -daemon")
	}
	setUmaskForDaemon()
	daemon.Serve(f.Sock, f.DB)
	return nil
}

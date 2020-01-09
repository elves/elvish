package program

import (
	"os"

	daemonsvc "github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/program/daemon"
)

// Daemon runs the daemon subprogram.
type Daemon struct {
	inner *daemon.Daemon
}

func (d Daemon) Main(fds [3]*os.File, _ []string) int {
	err := d.inner.Main(daemonsvc.Serve)
	if err != nil {
		logger.Println("daemon error:", err)
		return 2
	}
	return 0
}

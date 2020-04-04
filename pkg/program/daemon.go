package program

import (
	"os"

	"github.com/elves/elvish/pkg/daemon"
)

type daemonProgram struct {
	SockPath string
	DbPath   string
}

func (p daemonProgram) Main(fds [3]*os.File, _ []string) int {
	setUmaskForDaemon()
	daemon.Serve(p.SockPath, p.DbPath)
	return 0
}

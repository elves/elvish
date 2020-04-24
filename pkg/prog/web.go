package prog

import (
	"os"

	"github.com/elves/elvish/pkg/prog/web"
)

type webProgram struct{}

func (w webProgram) ShouldRun(f *Flags) bool { return f.Web }

func (w webProgram) Run(fds [3]*os.File, f *Flags, args []string) error {
	if len(args) > 0 {
		return BadUsage("arguments are not allowed with -web")
	}
	if f.CodeInArg {
		return BadUsage("-c cannot be used together with -web")
	}
	p := web.Web{BinPath: f.Bin, SockPath: f.Sock, DbPath: f.DB, Port: f.Port}
	return p.Main(fds, nil)
}

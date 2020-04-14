package program

import (
	"os"

	"github.com/elves/elvish/pkg/program/shell"
)

type shellProgram struct {
	BinPath     string
	SockPath    string
	DbPath      string
	Cmd         bool
	CompileOnly bool
	NoRc        bool
	JSON        bool
}

func (sh *shellProgram) Main(fds [3]*os.File, args []string) int {
	p := shell.MakePathsWithDefaults(fds[2],
		&shell.Paths{Bin: sh.BinPath, Sock: sh.SockPath, Db: sh.DbPath})
	if sh.NoRc {
		p.Rc = ""
	}
	if len(args) > 0 {
		return shell.Script(
			fds, args, &shell.ScriptConfig{
				SpawnDaemon: true, Paths: *p,
				Cmd: sh.Cmd, CompileOnly: sh.CompileOnly, JSON: sh.JSON})
	}
	shell.Interact(fds, &shell.InteractConfig{SpawnDaemon: true, Paths: *p})
	return 0
}

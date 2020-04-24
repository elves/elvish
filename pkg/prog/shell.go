package prog

import (
	"os"

	"github.com/elves/elvish/pkg/prog/shell"
)

type shellProgram struct{}

func (sh shellProgram) ShouldRun(*Flags) bool { return true }

func (sh shellProgram) Run(fds [3]*os.File, f *Flags, args []string) error {
	p := shell.MakePaths(fds[2],
		shell.Paths{Bin: f.Bin, Sock: f.Sock, Db: f.DB})
	if f.NoRc {
		p.Rc = ""
	}
	if len(args) > 0 {
		exit := shell.Script(
			fds, args, &shell.ScriptConfig{
				SpawnDaemon: true, Paths: p,
				Cmd: f.CodeInArg, CompileOnly: f.CompileOnly, JSON: f.JSON})
		return Exit(exit)
	}
	shell.Interact(fds, &shell.InteractConfig{SpawnDaemon: true, Paths: p})
	return nil
}

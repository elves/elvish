// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"fmt"
	"os"
	"time"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
)

var logger = logutil.GetLogger("[shell] ")

// Program is the shell subprogram.
type Program struct {
	ActivateDaemon daemondefs.ActivateFunc
}

func (p Program) ShouldRun(*prog.Flags) bool { return true }

func (p Program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	ev, cleanup1 := InitEvaler(fds[2])
	defer cleanup1()
	cleanup2 := initTTYAndSignal(fds[2])
	defer cleanup2()

	if len(args) > 0 {
		exit := Script(
			fds, args, &ScriptConfig{
				Evaler: ev,
				Cmd:    f.CodeInArg, CompileOnly: f.CompileOnly, JSON: f.JSON})
		return prog.Exit(exit)
	}

	spawnCfg, err := daemonPaths(f)
	if err != nil {
		fmt.Fprintln(fds[2], "Warning:", err)
		fmt.Fprintln(fds[2], "Storage daemon may not function.")
	}
	rc := ""
	if !f.NoRc {
		var err error
		rc, err = RCPath()
		if err != nil {
			fmt.Fprintln(fds[2], "Warning:", err)
		}
	}
	Interact(fds, &InteractConfig{
		Evaler: ev, RC: rc,
		ActivateDaemon: p.ActivateDaemon, SpawnConfig: spawnCfg})
	return nil
}

func evalInTTY(ev *eval.Evaler, fds [3]*os.File, src parse.Source) (float64, error) {
	start := time.Now()
	ports, cleanup := eval.PortsFromFiles(fds, ev.ValuePrefix())
	defer cleanup()
	err := ev.Eval(src, eval.EvalCfg{
		Ports: ports, Interrupt: eval.ListenInterrupts, PutInFg: true})
	end := time.Now()
	return end.Sub(start).Seconds(), err
}

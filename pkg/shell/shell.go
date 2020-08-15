// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"os"
	"os/signal"
	"strconv"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/sys"
	"github.com/elves/elvish/pkg/util"
)

var logger = util.GetLogger("[shell] ")

// Program is the shell subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) ShouldRun(*prog.Flags) bool { return true }

func (program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	p := MakePaths(fds[2],
		Paths{Bin: f.Bin, Sock: f.Sock, Db: f.DB})
	if f.NoRc {
		p.Rc = ""
	}
	if len(args) > 0 {
		exit := Script(
			fds, args, &ScriptConfig{
				Paths: p,
				Cmd:   f.CodeInArg, CompileOnly: f.CompileOnly, JSON: f.JSON})
		return prog.Exit(exit)
	}
	Interact(fds, &InteractConfig{SpawnDaemon: true, Paths: p})
	return nil
}

func setupShell(fds [3]*os.File, p Paths, spawn bool) (*eval.Evaler, func()) {
	restoreTTY := term.SetupGlobal()
	ev := InitRuntime(fds[2], p, spawn)
	restoreSHLVL := incSHLVL()
	sigCh := sys.NotifySignals()

	go func() {
		for sig := range sigCh {
			logger.Println("signal", sig)
			handleSignal(sig, fds[2])
		}
	}()

	return ev, func() {
		signal.Stop(sigCh)
		restoreSHLVL()
		CleanupRuntime(fds[2], ev)
		restoreTTY()
	}
}

func evalInTTY(ev *eval.Evaler, op eval.Op, fds [3]*os.File) error {
	ports, cleanup := eval.PortsFromFiles(fds, ev)
	defer cleanup()
	return ev.Eval(op, eval.EvalCfg{
		Ports: ports[:], Interrupt: eval.ListenInterrupts, PutInFg: true})
}

func incSHLVL() func() {
	restoreSHLVL := saveEnv(util.EnvSHLVL)

	i, err := strconv.Atoi(os.Getenv(util.EnvSHLVL))
	if err != nil {
		i = 0
	}
	os.Setenv(util.EnvSHLVL, strconv.Itoa(i+1))

	return restoreSHLVL
}

func saveEnv(name string) func() {
	v, ok := os.LookupEnv(name)
	if ok {
		return func() { os.Setenv(name, v) }
	}
	return func() { os.Unsetenv(name) }
}

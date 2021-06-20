// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"os"
	"os/signal"
	"strconv"
	"time"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/sys"
)

var logger = logutil.GetLogger("[shell] ")

// Program is the shell subprogram.
type Program struct {
	ActivateDaemon daemondefs.ActivateFunc
}

func (p Program) ShouldRun(*prog.Flags) bool { return true }

func (p Program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	paths := MakePaths(fds[2], Paths{Bin: f.Bin, Sock: f.Sock, Db: f.DB})
	if f.NoRc {
		paths.Rc = ""
	}
	if len(args) > 0 {
		exit := Script(
			fds, args, &ScriptConfig{
				Paths: paths,
				Cmd:   f.CodeInArg, CompileOnly: f.CompileOnly, JSON: f.JSON})
		return prog.Exit(exit)
	}
	Interact(fds, &InteractConfig{ActivateDaemon: p.ActivateDaemon, Paths: paths})
	return nil
}

func setupShell(fds [3]*os.File, p Paths, activate daemondefs.ActivateFunc) (*eval.Evaler, func()) {
	restoreTTY := term.SetupGlobal()
	ev := InitRuntime(fds[2], p, activate)
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

func evalInTTY(ev *eval.Evaler, fds [3]*os.File, src parse.Source) (float64, error) {
	start := time.Now()
	ports, cleanup := eval.PortsFromFiles(fds, ev.ValuePrefix())
	defer cleanup()
	err := ev.Eval(src, eval.EvalCfg{
		Ports: ports, Interrupt: eval.ListenInterrupts, PutInFg: true})
	end := time.Now()
	return end.Sub(start).Seconds(), err
}

func incSHLVL() func() {
	restoreSHLVL := saveEnv(env.SHLVL)

	i, err := strconv.Atoi(os.Getenv(env.SHLVL))
	if err != nil {
		i = 0
	}
	os.Setenv(env.SHLVL, strconv.Itoa(i+1))

	return restoreSHLVL
}

func saveEnv(name string) func() {
	v, ok := os.LookupEnv(name)
	if ok {
		return func() { os.Setenv(name, v) }
	}
	return func() { os.Unsetenv(name) }
}

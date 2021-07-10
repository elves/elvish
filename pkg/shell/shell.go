// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"time"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/mods/file"
	mathmod "src.elv.sh/pkg/eval/mods/math"
	pathmod "src.elv.sh/pkg/eval/mods/path"
	"src.elv.sh/pkg/eval/mods/platform"
	"src.elv.sh/pkg/eval/mods/re"
	"src.elv.sh/pkg/eval/mods/str"
	"src.elv.sh/pkg/eval/mods/unix"
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
	cleanup1 := IncSHLVL()
	defer cleanup1()
	cleanup2 := initTTYAndSignal(fds[2])
	defer cleanup2()

	ev := MakeEvaler(fds[2])

	if len(args) > 0 {
		exit := script(
			ev, fds, args, &scriptCfg{
				Cmd: f.CodeInArg, CompileOnly: f.CompileOnly, JSON: f.JSON})
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
		rc, err = rcPath()
		if err != nil {
			fmt.Fprintln(fds[2], "Warning:", err)
		}
	}
	interact(ev, fds, &interactCfg{
		RC:             rc,
		ActivateDaemon: p.ActivateDaemon, SpawnConfig: spawnCfg})
	return nil
}

// MakeEvaler creates an Evaler, sets the module search directories and installs
// all the standard builtin modules. It writes a warning message to the supplied
// Writer if it could not initialize module search directories.
func MakeEvaler(stderr io.Writer) *eval.Evaler {
	ev := eval.NewEvaler()
	libs, libInstall, err := libPaths()
	if err != nil {
		fmt.Fprintln(stderr, "Warning:", err)
	}
	ev.SetLibDirs(libs)
	ev.SetLibInstallDir(libInstall)
	ev.AddModule("math", mathmod.Ns)
	ev.AddModule("path", pathmod.Ns)
	ev.AddModule("platform", platform.Ns)
	ev.AddModule("re", re.Ns)
	ev.AddModule("str", str.Ns)
	ev.AddModule("file", file.Ns)
	if unix.ExposeUnixNs {
		ev.AddModule("unix", unix.Ns)
	}
	return ev
}

// IncSHLVL increments the SHLVL environment variable. It returns a function to
// restore the original value of SHLVL.
func IncSHLVL() func() {
	oldValue, hadValue := os.LookupEnv(env.SHLVL)
	i, err := strconv.Atoi(oldValue)
	if err != nil {
		i = 0
	}
	os.Setenv(env.SHLVL, strconv.Itoa(i+1))

	if hadValue {
		return func() { os.Setenv(env.SHLVL, oldValue) }
	} else {
		return func() { os.Unsetenv(env.SHLVL) }
	}
}

func initTTYAndSignal(stderr io.Writer) func() {
	restoreTTY := term.SetupGlobal()

	sigCh := sys.NotifySignals()
	go func() {
		for sig := range sigCh {
			logger.Println("signal", sig)
			handleSignal(sig, stderr)
		}
	}()

	return func() {
		signal.Stop(sigCh)
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

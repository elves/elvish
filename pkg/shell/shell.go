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
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/mods"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/sys"
)

var logger = logutil.GetLogger("[shell] ")

// Program is the shell subprogram.
type Program struct {
	ActivateDaemon daemondefs.ActivateFunc

	codeInArg   bool
	compileOnly bool
	noRC        bool
	rc          string
	json        *bool
	daemonPaths *prog.DaemonPaths
}

func (p *Program) RegisterFlags(fs *prog.FlagSet) {
	// Support -i so that programs that expect shells to support it (like
	// "script") don't error when they invoke Elvish.
	fs.Bool("i", false,
		"A no-op flag, introduced for POSIX compatibility")
	fs.BoolVar(&p.codeInArg, "c", false,
		"Treat the first argument as code to execute")
	fs.BoolVar(&p.compileOnly, "compileonly", false,
		"Parse and compile Elvish code without executing it")
	fs.BoolVar(&p.noRC, "norc", false,
		"Don't read the RC file when running interactively")
	fs.StringVar(&p.rc, "rc", "",
		"Path to the RC file when running interactively")

	p.json = fs.JSON()
	if p.ActivateDaemon != nil {
		p.daemonPaths = fs.DaemonPaths()
	}
}

func (p *Program) Run(fds [3]*os.File, args []string) error {
	cleanup1 := IncSHLVL()
	defer cleanup1()
	cleanup2 := initSignal(fds)
	defer cleanup2()

	ev := MakeEvaler(fds[2])

	if len(args) > 0 {
		exit := script(
			ev, fds, args, &scriptCfg{
				Cmd: p.codeInArg, CompileOnly: p.compileOnly, JSON: *p.json})
		return prog.Exit(exit)
	}

	var spawnCfg *daemondefs.SpawnConfig
	if p.ActivateDaemon != nil {
		var err error
		spawnCfg, err = daemonPaths(p.daemonPaths)
		if err != nil {
			fmt.Fprintln(fds[2], "Warning:", err)
			fmt.Fprintln(fds[2], "Storage daemon may not function.")
		}
	}

	rc := ""
	switch {
	case p.noRC:
		// Leave rc empty
	case p.rc != "":
		// Use explicit -rc flag value
		rc = p.rc
	default:
		// Use default path to rc.elv
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
	libs, err := libPaths()
	if err != nil {
		fmt.Fprintln(stderr, "Warning:", err)
	}
	ev.LibDirs = libs
	mods.AddTo(ev)
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

func initSignal(fds [3]*os.File) func() {
	sigCh := sys.NotifySignals()
	go func() {
		for sig := range sigCh {
			logger.Println("signal", sig)
			handleSignal(sig, fds[2])
		}
	}()

	return func() { signal.Stop(sigCh) }
}

func evalInTTY(fds [3]*os.File, ev *eval.Evaler, ed editor, src parse.Source) error {
	start := time.Now()
	ports, cleanup := eval.PortsFromFiles(fds, ev.ValuePrefix())
	defer cleanup()
	restore := term.SetupForEval(fds[0], fds[1])
	defer restore()
	err := ev.Eval(src, eval.EvalCfg{
		Ports: ports, Interrupt: eval.ListenInterrupts, PutInFg: true})
	if ed != nil {
		ed.RunAfterCommandHooks(src, time.Since(start).Seconds(), err)
	}
	return err
}

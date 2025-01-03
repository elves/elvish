// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
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
	"src.elv.sh/pkg/ui"
)

var logger = logutil.GetLogger("[shell] ")

// Program is the shell subprogram. It has two slightly different modes of
// operation:
//
//   - When the command line argument contains a filename or "-c some-code", the
//     shell is non-interactive. In this mode, it just evaluates the given file or
//     code.
//
//   - Otherwise, the shell is interactive, and launches a terminal [REPL]. This
//     mode also initializes the storage backend, which in turn activates the
//     storage daemon.
//
//     To enable building a daemon-less version, the subprogram doesn't depend
//     on pkg/daemon, and the caller should supply pkg/daemon.Activate in the
//     ActivateDaemon field to enable functionalities. If it is nil, daemon
//     functionalities are disabled.
//
// [REPL]: https://en.wikipedia.org/wiki/Read–eval–print_loop
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
	// script(1) (and possibly other programs) assume shells support -i
	fs.Bool("i", false,
		"A no-op flag, introduced for compatibility")
	// termux (and possibly other programs) assume shells support -l
	fs.Bool("l", false,
		"A no-op flag, introduced for compatibility")
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
	cleanup1 := incSHLVL()
	defer cleanup1()
	cleanup2 := initSignal(fds)
	defer cleanup2()

	// https://no-color.org
	ui.NoColor = os.Getenv(env.NO_COLOR) != ""
	interactive := len(args) == 0
	ev := p.makeEvaler(fds[2], interactive)
	defer ev.PreExit()

	if !interactive {
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

	interact(ev, fds, &interactCfg{
		RC:             ev.EffectiveRcPath,
		ActivateDaemon: p.ActivateDaemon, SpawnConfig: spawnCfg})
	return nil
}

// Creates an Evaler, sets the module search directories and installs all the
// standard builtin modules.
//
// It writes a warning message to the supplied Writer if it could not initialize
// module search directories.
func (p *Program) makeEvaler(stderr io.Writer, interactive bool) *eval.Evaler {
	ev := eval.NewEvaler()

	var errRc error
	ev.RcPath, errRc = rcPath()
	switch {
	case !interactive || p.noRC:
		// Leave ev.ActualRcPath empty
	case p.rc != "":
		// Use explicit -rc flag value
		var err error
		ev.EffectiveRcPath, err = filepath.Abs(p.rc)
		if err != nil {
			fmt.Fprintln(stderr, "Warning:", err)
		}
	default:
		if errRc == nil {
			// Use default path stored in ev.RcPath
			ev.EffectiveRcPath = ev.RcPath
		} else {
			fmt.Fprintln(stderr, "Warning:", errRc)
		}
	}

	libs, err := libPaths()
	if err != nil {
		fmt.Fprintln(stderr, "Warning: resolving lib paths:", err)
	} else {
		ev.LibDirs = libs
	}

	mods.AddTo(ev)
	return ev
}

// Increments the SHLVL environment variable. It returns a function to restore
// the original value of SHLVL.
func incSHLVL() func() {
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

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}
}

func evalInTTY(fds [3]*os.File, ev *eval.Evaler, ed editor, src parse.Source) error {
	start := time.Now()
	ports, cleanup := eval.PortsFromFiles(fds, ev.ValuePrefix())
	defer cleanup()
	restore := term.SetupForEval(fds[0], fds[1])
	defer restore()
	ctx, done := eval.ListenInterrupts()
	err := ev.Eval(src, eval.EvalCfg{
		Ports: ports, Interrupts: ctx, PutInFg: true})
	done()
	if ed != nil {
		ed.RunAfterCommandHooks(src, time.Since(start).Seconds(), err)
	}
	return err
}

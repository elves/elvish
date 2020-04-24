// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/sys"
	"github.com/elves/elvish/pkg/util"
)

var logger = util.GetLogger("[shell] ")

func setupShell(fds [3]*os.File, p Paths, spawn bool) (*eval.Evaler, func()) {
	restoreTTY := term.SetupGlobal()
	ev := InitRuntime(fds[2], p, spawn)
	restoreSHLVL := incSHLVL()

	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			logger.Println("signal", sig)
			handleSignal(sig, fds[2])
		}
	}()

	return ev, func() {
		signal.Stop(sigs)
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

const envSHLVL = "SHLVL"

func incSHLVL() func() {
	restoreSHLVL := saveEnv(envSHLVL)

	i, err := strconv.Atoi(os.Getenv(envSHLVL))
	if err != nil {
		i = 0
	}
	os.Setenv(envSHLVL, strconv.Itoa(i+1))

	return restoreSHLVL
}

func saveEnv(name string) func() {
	v, ok := os.LookupEnv(name)
	if ok {
		return func() { os.Setenv(name, v) }
	}
	return func() { os.Unsetenv(name) }
}

// Global panic handler.
func rescue() {
	r := recover()
	if r != nil {
		println()
		print(sys.DumpStack())
		println()
		println(r)
		println("\nExecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

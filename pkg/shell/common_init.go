package shell

import (
	"io"
	"os"
	"os/signal"
	"strconv"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/mods/file"
	mathmod "src.elv.sh/pkg/eval/mods/math"
	pathmod "src.elv.sh/pkg/eval/mods/path"
	"src.elv.sh/pkg/eval/mods/platform"
	"src.elv.sh/pkg/eval/mods/re"
	"src.elv.sh/pkg/eval/mods/str"
	"src.elv.sh/pkg/eval/mods/unix"
	"src.elv.sh/pkg/sys"
)

// InitEvaler creates an Evaler, sets the search directory for modules, installs
// all the standard builtin modules, and increases SHLVL. It returns the Evaler
// and a callback to restore the old SHLVL.
func InitEvaler(libDir string) (*eval.Evaler, func()) {
	ev := eval.NewEvaler()
	ev.SetLibDir(libDir)
	ev.AddModule("math", mathmod.Ns)
	ev.AddModule("path", pathmod.Ns)
	ev.AddModule("platform", platform.Ns)
	ev.AddModule("re", re.Ns)
	ev.AddModule("str", str.Ns)
	ev.AddModule("file", file.Ns)
	if unix.ExposeUnixNs {
		ev.AddModule("unix", unix.Ns)
	}
	restoreSHLVL := incSHLVL()
	return ev, restoreSHLVL
}

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

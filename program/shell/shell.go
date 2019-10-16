// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/runtime"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[shell] ")

// Shell keeps flags to the shell.
type Shell struct {
	BinPath     string
	SockPath    string
	DbPath      string
	Cmd         bool
	CompileOnly bool
	NoRc        bool
	NewEdit     bool
	JSON        bool
}

func New(binpath, sockpath, dbpath string, cmd, compileonly, norc, newEdit, json bool) *Shell {
	return &Shell{binpath, sockpath, dbpath, cmd, compileonly, norc, newEdit, json}
}

// Main runs Elvish using the default terminal interface. It blocks until Elvish
// quits, and returns the exit code.
func (sh *Shell) Main(args []string) int {
	defer rescue()

	restoreTTY := term.SetupGlobal()
	defer restoreTTY()

	ev, dataDir := runtime.InitRuntime(sh.BinPath, sh.SockPath, sh.DbPath)
	defer runtime.CleanupRuntime(ev)

	handleSignals()

	if len(args) > 0 {
		err := script(ev, args, sh.Cmd, sh.CompileOnly)
		if err != nil {
			if sh.CompileOnly && sh.JSON {
				fmt.Println(ErrorToJSON(err))
			} else {
				diag.PPrintError(err)
			}
			return 2
		}
	} else {
		interact(ev, dataDir, sh.NoRc, sh.NewEdit)
	}

	return 0
}

// Global panic handler.
func rescue() {
	r := recover()
	if r != nil {
		println()
		fmt.Println(r)
		print(sys.DumpStack())
		println("\nexecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

func handleSignals() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			logger.Println("signal", sig)
			handleSignal(sig)
		}
	}()
}

// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
}

func New(binpath, sockpath, dbpath string, cmd, compileonly, norc bool) *Shell {
	return &Shell{binpath, sockpath, dbpath, cmd, compileonly, norc}
}

// Main runs Elvish using the default terminal interface. It blocks until Elvish
// quits, and returns the exit code.
func (sh *Shell) Main(args []string) int {
	defer rescue()

	ev, dataDir := runtime.InitRuntime(sh.BinPath, sh.SockPath, sh.DbPath)
	defer runtime.CleanupRuntime(ev)

	handleSignals()

	if len(args) > 0 {
		err := script(ev, args, sh.Cmd, sh.CompileOnly)
		if err != nil {
			util.PprintError(err)
			return 2
		}
	} else {
		interact(ev, dataDir, sh.NoRc)
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

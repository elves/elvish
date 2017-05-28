// Elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"syscall"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/shell"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// closeFd is used in syscall.ProcAttr.Files to signify closing a fd.
const closeFd = ^uintptr(0)

var (
	// Flags handled in this package, or common to shell and daemon.
	help = flag.Bool("help", false, "show usage help and quit")

	logpath    = flag.String("log", "", "a file to write debug log to")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	dbpath     = flag.String("db", "", "path to the database")
	sockpath   = flag.String("sock", "", "path to the daemon socket")

	isdaemon = flag.Bool("daemon", false, "run daemon instead of shell")
	isweb    = flag.Bool("web", false, "run backend of web interface")

	// Flags for shell and web.
	cmd = flag.Bool("c", false, "take first argument as a command to execute")

	// Flags for daemon.
	forked  = flag.Int("forked", 0, "how many times the daemon has forked")
	binpath = flag.String("bin", "", "path to the elvish binary")
)

func usage() {
	fmt.Println("usage: elvish [flags] [script]")
	fmt.Println("flags:")
	flag.PrintDefaults()
}

func main() {
	// This is needed for defers to be honored.
	ret := 0
	defer os.Exit(ret)

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if *help {
		usage()
		return
	}

	// The daemon takes no argument.
	if *isdaemon && len(args) > 0 {
		usage()
		ret = 2
		return
	}

	if *logpath != "" {
		err := util.SetOutputFile(*logpath)
		if err != nil {
			fmt.Println(err)
		}
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *isdaemon {
		ret = doDaemon()
	} else {
		// Set up Evaler and Store. This is needed for both the terminal and web
		// interfaces.
		ev, st := newEvalerAndStore(*dbpath)
		defer func() {
			err := st.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to close database:", err)
			}
		}()

		if *isweb {
			fmt.Fprintln(os.Stderr, "web interface not yet implemented.")
			ret = 2
		} else {
			sh := shell.NewShell(ev, st, *cmd)
			ret = sh.Run(args)
		}
	}
}

func newEvalerAndStore(db string) (*eval.Evaler, *store.Store) {
	dataDir, err := store.EnsureDataDir()

	var st *store.Store
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: cannot create data dir ~/.elvish")
	} else {
		if db == "" {
			db = dataDir + "/db"
		}
		st, err = store.NewStore(db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: cannot connect to store:", err)
		}
	}

	return eval.NewEvaler(st, dataDir), st
}

func doDaemon() int {
	switch *forked {
	case 0:
		errored := false
		absify := func(f string, s *string) {
			if *s == "" {
				log.Println("flag", f, "is required for daemon")
				errored = true
				return
			}
			p, err := filepath.Abs(*s)
			if err != nil {
				log.Println("abs:", err)
				errored = true
			} else {
				*s = p
			}
		}
		absify("-bin", binpath)
		absify("-db", dbpath)
		absify("-sock", sockpath)
		absify("-log", logpath)
		if errored {
			return 2
		}

		syscall.Umask(0077)
		return forkDaemon(
			&syscall.ProcAttr{
				// cd to /
				Dir: "/",
				// empty environment
				Env: nil,
				// inherit stderr only for logging
				Files: []uintptr{closeFd, closeFd, 2},
				Sys:   &syscall.SysProcAttr{Setsid: true},
			})
	case 1:
		return forkDaemon(
			&syscall.ProcAttr{
				Files: []uintptr{closeFd, closeFd, 2},
			})
	case 2:
		d := daemon.New(*sockpath, *dbpath)
		return d.Main()
	default:
		return 2
	}
}

func forkDaemon(attr *syscall.ProcAttr) int {
	_, err := syscall.ForkExec(*binpath, []string{
		*binpath,
		"-daemon",
		"-forked", strconv.Itoa(*forked + 1),
		"-bin", *binpath,
		"-db", *dbpath,
		"-sock", *sockpath,
		"-log", *logpath,
	}, attr)
	if err != nil {
		log.Println("fork/exec:", err)
		return 2
	}
	return 0
}

// Elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/shell"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
	"github.com/elves/elvish/web"
)

// closeFd is used in syscall.ProcAttr.Files to signify closing a fd.
const closeFd = ^uintptr(0)

// defaultPort is the default port on which the web interface runs. The number
// is chosen because it resembles "elvi".
const defaultWebPort = 3171

var logger = util.GetLogger("[main] ")

var (
	// Flags handled in this package, or common to shell and daemon.
	help = flag.Bool("help", false, "show usage help and quit")

	logpath    = flag.String("log", "", "a file to write debug log to")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	dbpath     = flag.String("db", "", "path to the database")
	sockpath   = flag.String("sock", "", "path to the daemon socket")

	isdaemon = flag.Bool("daemon", false, "run daemon instead of shell")
	isweb    = flag.Bool("web", false, "run backend of web interface")
	webport  = flag.Int("port", defaultWebPort, "the port of the web backend")

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

	// Parse and check flags.
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if *help {
		usage()
		return
	}
	if *isdaemon && len(args) > 0 {
		// The daemon takes no argument.
		usage()
		ret = 2
		return
	}

	// Flags common to all sub-programs: log and CPU profile.
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

	// Pick a sub-program to run.
	if *isdaemon {
		ret = doDaemon()
	} else {
		// Shell or web. Set up common runtime components.
		ev, st, client := initRuntime()
		if client != nil {
			req := &api.PidRequest{}
			res := &api.PidResponse{}
			go func() {
				client.Call(api.ServiceName+".Pid", req, res)
				logger.Println("daemon pid is", res.Pid)
			}()
		}
		defer func() {
			err := st.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: failed to close database:", err)
			}
			err = client.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: failed to close connection to daemon:", err)
			}
		}()

		if *isweb {
			if *cmd {
				fmt.Fprintln(os.Stderr, "-c -web not yet supported")
				ret = 2
				return
			}
			w := web.NewWeb(ev, st, *webport)
			ret = w.Run(args)
		} else {
			sh := shell.NewShell(ev, st, *cmd)
			ret = sh.Run(args)
		}
	}
}

func initRuntime() (*eval.Evaler, *store.Store, *rpc.Client) {
	var dataDir string
	var err error
	if *dbpath == "" || *sockpath == "" {
		// Determine default paths for database and socket.
		dataDir, err = store.EnsureDataDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot create data dir ~/.elvish")
		} else {
			if *dbpath == "" {
				*dbpath = dataDir + "/db"
			}
			if *sockpath == "" {
				*sockpath = dataDir + "/sock"
			}
		}
	}

	var st *store.Store
	if *dbpath != "" {
		st, err = store.NewStore(*dbpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot connect to store:", err)
		}
	}

	var client *rpc.Client
	if *sockpath != "" && *dbpath != "" {
		if _, err := os.Stat(*sockpath); os.IsNotExist(err) {
			logger.Println("socket does not exists, starting daemon")
			err := startDaemon(dataDir + "/daemon.log")
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: cannot start daemon:", err)
			}
			goto endCreateClient
		}
		client, err = rpc.Dial("unix", *sockpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot connect to daemon:", err)
		}
	}
endCreateClient:

	return eval.NewEvaler(st, client, dataDir), st, client
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

var errCannotDetermineBinPath = errors.New("cannot determine path of elvish binary")

// startDaemon starts a daemon in the background. It is supposed to be called
// from a client.
func startDaemon(logpath string) error {
	// Determine binpath.
	if *binpath == "" {
		if len(os.Args) > 0 && path.IsAbs(os.Args[0]) {
			*binpath = os.Args[0]
		} else {
			// Find elvish in PATH
			return errCannotDetermineBinPath
			paths := strings.Split(os.Getenv("PATH"), ":")
			bin, err := util.Search(paths, "elvish")
			if err != nil {
				return errors.New("cannot find elvish: " + err.Error())
			}
			*binpath = bin
		}
	}
	return runDaemon(nil, 0, logpath)
}

// forkDaemon forks a daemon. It is supposed to be called from the daemon.
func forkDaemon(attr *syscall.ProcAttr) int {
	err := runDaemon(attr, *forked+1, *logpath)
	if err != nil {
		return 2
	}
	return 0
}

func runDaemon(attr *syscall.ProcAttr, forklevel int, logpath string) error {
	_, err := syscall.ForkExec(*binpath, []string{
		*binpath,
		"-daemon",
		"-forked", strconv.Itoa(forklevel),
		"-bin", *binpath,
		"-db", *dbpath,
		"-sock", *sockpath,
		"-log", logpath,
	}, attr)
	return err
}

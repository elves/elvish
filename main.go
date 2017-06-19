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
	"runtime/pprof"
	"strconv"
	"syscall"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/daemon/client"
	"github.com/elves/elvish/daemon/service"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/shell"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
	"github.com/elves/elvish/web"
)

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
	forked        = flag.Int("forked", 0, "how many times the daemon has forked")
	binpath       = flag.String("bin", "", "path to the elvish binary")
	logpathprefix = flag.String("logprefix", "", "the prefix for the daemon log file")
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
	if *isdaemon {
		if *forked == 2 && *logpathprefix != "" {
			// Honor logpathprefix.
			pid := syscall.Getpid()
			err := util.SetOutputFile(*logpathprefix + strconv.Itoa(pid))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			util.SetOutputFile("/dev/stderr")
		}
	} else if *logpath != "" {
		err := util.SetOutputFile(*logpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
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
		d := daemon.Daemon{
			Forked:        *forked,
			BinPath:       *binpath,
			DbPath:        *dbpath,
			SockPath:      *sockpath,
			LogPathPrefix: *logpathprefix,
		}
		ret = d.Main(service.Serve)
	} else {
		// Shell or web. Set up common runtime components.
		ev, st, cl := initRuntime()
		defer func() {
			err := st.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: failed to close database:", err)
			}
			err = cl.Close()
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

func initRuntime() (*eval.Evaler, *store.Store, *client.Client) {
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

	toSpawn := &daemon.Daemon{
		Forked:        *forked,
		BinPath:       *binpath,
		DbPath:        *dbpath,
		SockPath:      *sockpath,
		LogPathPrefix: dataDir + "/daemon.log.",
	}
	var cl *client.Client
	if *sockpath != "" && *dbpath != "" {
		cl = client.New(*sockpath)
		_, statErr := os.Stat(*sockpath)
		killed := false
		if statErr == nil {
			// Kill the daemon if it is outdated.
			req := &api.VersionRequest{}
			res := &api.VersionResponse{}
			err := cl.CallDaemon("Version", req, res)
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: socket exists but not responding version RPC:", err)
				goto spawnDaemonEnd
			}
			logger.Printf("daemon serving version %d, want version %d", res.Version, api.Version)
			if res.Version < api.Version {
				req := &api.PidRequest{}
				res := &api.PidResponse{}
				err := cl.CallDaemon("Pid", req, res)
				if err != nil {
					fmt.Fprintln(os.Stderr, "warning: socket exists but not responding pid RPC:", err)
					goto spawnDaemonEnd
				}
				logger.Printf("killing outdated daemon with pid %d", res.Pid)
				err = syscall.Kill(res.Pid, syscall.SIGTERM)
				if err != nil {
					fmt.Fprintln(os.Stderr, "warning: failed to kill outdated daemon process:", err)
					goto spawnDaemonEnd
				}
				fmt.Fprintln(os.Stderr, "killed outdated daemon")
				killed = true
			}
		}
		if os.IsNotExist(statErr) || killed {
			logger.Println("socket does not exists, starting daemon")
			err := toSpawn.Spawn()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: cannot start daemon:", err)
			} else {
				fmt.Fprintln(os.Stderr, "started daemon")
			}
		}
	}
spawnDaemonEnd:

	return eval.NewEvaler(st, cl, toSpawn, dataDir), st, cl
}

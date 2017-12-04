// Package program provides the entry point to Elvish. Its subpackages
// correspond to subprograms of Elvish.
package program

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/elves/elvish/build"
	daemonsvc "github.com/elves/elvish/daemon"
	daemonapi "github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/re"
	"github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/program/shell"
	"github.com/elves/elvish/program/web"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

// defaultPort is the default port on which the web interface runs. The number
// is chosen because it resembles "elvi".
const defaultWebPort = 3171

var logger = util.GetLogger("[main] ")

var (
	// Flags handled in this package, or common to shell and daemon.
	help        = flag.Bool("help", false, "show usage help and quit")
	showVersion = flag.Bool("version", false, "show version and quit")

	logpath     = flag.String("log", "", "a file to write debug log to")
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file")
	dbpath      = flag.String("db", "", "path to the database")
	sockpath    = flag.String("sock", "", "path to the daemon socket")
	compileonly = flag.Bool("compileonly", false, "Parse/Compile but do not execute")

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

func Main() int {
	// Parse and check flags.
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if *help {
		usage()
		return 0
	}

	if *showVersion {
		fmt.Println(build.Version)
		return 0
	}

	if *isdaemon && len(args) > 0 {
		// The daemon takes no argument.
		usage()
		return 2
	}

	// Flags common to all sub-programs: log and CPU profile.
	if *isdaemon {
		if *forked == 2 && *logpathprefix != "" {
			// Honor logpathprefix.
			pid := os.Getpid()
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
		err := d.Main(daemonsvc.Serve)
		if err != nil {
			logger.Println("daemon -forked", *forked, "error:", err)
			return 2
		}
		return 0
	} else {
		// Shell or web. Set up common runtime components.
		ev, cl := initRuntime()
		if cl != nil {
			defer func() {
				err := cl.Close()
				if err != nil {
					fmt.Fprintln(os.Stderr, "warning: failed to close connection to daemon:", err)
				}
			}()
		}

		if *isweb {
			if *cmd {
				fmt.Fprintln(os.Stderr, "-c -web not yet supported")
				return 2
			}
			w := web.NewWeb(ev, *webport)
			return w.Run(args)
		} else {
			sh := shell.NewShell(ev, cl, *cmd, *compileonly)
			return sh.Run(args)
		}
	}
}

const (
	daemonWaitOneLoop = 10 * time.Millisecond
	daemonWaitLoops   = 100
	daemonWaitTotal   = daemonWaitOneLoop * daemonWaitLoops
)

const upgradeDbNotice = `If you upgraded Elvish from a pre-0.10 version, you need to upgrade your database by following instructions in https://github.com/elves/upgrade-db-for-0.10/`

func initRuntime() (*eval.Evaler, *daemonapi.Client) {
	var dataDir string
	var err error

	// Determine data directory.
	dataDir, err = storedefs.EnsureDataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: cannot create data directory ~/.elvish")
	} else {
		if *dbpath == "" {
			*dbpath = dataDir + "/db"
		}
	}

	// Determine runtime directory.
	runDir, err := getSecureRunDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot get runtime dir /tmp/elvish-$uid, falling back to data dir ~/.elvish:", err)
		runDir = dataDir
	}
	if *sockpath == "" {
		*sockpath = runDir + "/sock"
	}

	toSpawn := &daemon.Daemon{
		Forked:        *forked,
		BinPath:       *binpath,
		DbPath:        *dbpath,
		SockPath:      *sockpath,
		LogPathPrefix: runDir + "/daemon.log.",
	}
	var cl *daemonapi.Client
	if *sockpath != "" && *dbpath != "" {
		cl = daemonapi.NewClient(*sockpath)
		_, statErr := os.Stat(*sockpath)
		killed := false
		if statErr == nil {
			// Kill the daemon if it is outdated.
			version, err := cl.Version()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: socket exists but not responding version RPC:", err)
				// TODO(xiaq): Remove this when the SQLite-backed database
				// becomes an unmemorable past (perhaps 6 months after the
				// switch to boltdb).
				if err.Error() == bolt.ErrInvalid.Error() {
					fmt.Fprintln(os.Stderr, upgradeDbNotice)
				}
				goto spawnDaemonEnd
			}
			logger.Printf("daemon serving version %d, want version %d", version, daemonapi.Version)
			if version < daemonapi.Version {
				pid, err := cl.Pid()
				if err != nil {
					fmt.Fprintln(os.Stderr, "warning: socket exists but not responding pid RPC:", err)
					cl.Close()
					cl = nil
					goto spawnDaemonEnd
				}
				cl.Close()
				logger.Printf("killing outdated daemon with pid %d", pid)
				p, err := os.FindProcess(pid)
				if err != nil {
					err = p.Kill()
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "warning: failed to kill outdated daemon process:", err)
					cl = nil
					goto spawnDaemonEnd
				}
				logger.Println("killed outdated daemon")
				killed = true
			}
		}
		if os.IsNotExist(statErr) || killed {
			logger.Println("socket does not exists, starting daemon")
			err := toSpawn.Spawn()
			if err != nil {
				fmt.Fprintln(os.Stderr, "warning: cannot start daemon:", err)
			} else {
				logger.Println("started daemon")
			}
			for i := 0; i <= daemonWaitLoops; i++ {
				_, err := cl.Version()
				if err == nil {
					logger.Println("daemon online")
					goto spawnDaemonEnd
				} else if err.Error() == bolt.ErrInvalid.Error() {
					fmt.Fprintln(os.Stderr, upgradeDbNotice)
					goto spawnDaemonEnd
				} else if i == daemonWaitLoops {
					fmt.Fprintf(os.Stderr, "cannot connect to daemon after %v: %v\n", daemonWaitTotal, err)
					goto spawnDaemonEnd
				}
				time.Sleep(daemonWaitOneLoop)
			}
		}
	}
spawnDaemonEnd:

	// TODO(xiaq): This information might belong somewhere else.
	extraModules := map[string]eval.Namespace{
		"re": re.Namespace(),
	}
	return eval.NewEvaler(cl, toSpawn, dataDir, extraModules), cl
}

var (
	ErrBadOwner      = errors.New("bad owner")
	ErrBadPermission = errors.New("bad permission")
)

// getSecureRunDir stats /tmp/elvish-$uid, creating it if it doesn't yet exist,
// and return the directory name if it has the correct owner and permission.
func getSecureRunDir() (string, error) {
	uid := os.Getuid()

	runDir := path.Join(os.TempDir(), fmt.Sprintf("elvish-%d", uid))
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	info, err := os.Stat(runDir)
	if err != nil {
		return "", err
	}

	return runDir, checkExclusiveAccess(info, uid)
}

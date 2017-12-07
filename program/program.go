// Package program provides the entry point to Elvish. Its subpackages
// correspond to subprograms of Elvish.
package program

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

	"github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/program/shell"
	"github.com/elves/elvish/program/web"
	"github.com/elves/elvish/util"
)

// defaultPort is the default port on which the web interface runs. The number
// is chosen because it resembles "elvi".
const defaultWebPort = 3171

var logger = util.GetLogger("[main] ")

var (
	// Flags handled in this package, or common to shell and daemon.
	help          = flag.Bool("help", false, "show usage help and quit")
	showVersion   = flag.Bool("version", false, "show version and quit")
	showBuildInfo = flag.Bool("buildinfo", false, "show build info and quit")
	showJSON      = flag.Bool("json", false, "show output in JSON. Useful with -buildinfo.")

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
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	// Handle flags common to all subprograms.

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var err error
	if *logpath != "" {
		err = util.SetOutputFile(*logpath)
	} else if *logpathprefix != "" {
		err = util.SetOutputFile(*logpathprefix + strconv.Itoa(os.Getpid()))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return FindProgram(args).Main(args)
}

// Program represents a subprogram.
type Program interface {
	// Main calls the subprogram with arguments. The return value will be used
	// as the exit status of the entire program.
	Main(args []string) int
}

// FindProgram finds a suitable Program according to flags. It does not have any
// side effects.
func FindProgram(args []string) Program {
	switch {
	case *help:
		return ShowHelp{}
	case *showVersion:
		return ShowVersion{}
	case *showBuildInfo:
		return ShowBuildInfo{*showJSON}
	case *isdaemon:
		if len(args) > 0 {
			// The daemon takes no argument.
			return ShowCorrectUsage{}
		}
		return Daemon{inner: &daemon.Daemon{
			Forked:        *forked,
			BinPath:       *binpath,
			DbPath:        *dbpath,
			SockPath:      *sockpath,
			LogPathPrefix: *logpathprefix,
		}}
	case *isweb:
		if *cmd || len(args) > 0 {
			return ShowCorrectUsage{}
		}
		return web.New(*binpath, *sockpath, *dbpath, *webport)
	default:
		return shell.New(*binpath, *sockpath, *dbpath, *cmd, *compileonly)
	}
}

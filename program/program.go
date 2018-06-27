// Package program provides the entry point to Elvish. Its subpackages
// correspond to subprograms of Elvish.
package program

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"flag"
	"fmt"
	"io"
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

type flagSet struct {
	flag.FlagSet

	Log, LogPrefix, CPUProfile string

	Help, Version, BuildInfo, JSON bool

	CodeInArg, CompileOnly, NoRc bool

	Web  bool
	Port int

	Daemon bool
	Forked int

	Bin, DB, Sock string
}

func newFlagSet() *flagSet {
	f := flagSet{}
	f.Init("elvish", flag.ContinueOnError)
	f.Usage = func() {
		usage(os.Stderr, &f)
	}

	f.StringVar(&f.Log, "log", "", "a file to write debug log to")
	f.StringVar(&f.LogPrefix, "logprefix", "", "the prefix for the daemon log file")
	f.StringVar(&f.CPUProfile, "cpuprofile", "", "write cpu profile to file")

	f.BoolVar(&f.Help, "help", false, "show usage help and quit")
	f.BoolVar(&f.Version, "version", false, "show version and quit")
	f.BoolVar(&f.BuildInfo, "buildinfo", false, "show build info and quit")
	f.BoolVar(&f.JSON, "json", false, "show output in JSON. Useful with -buildinfo.")

	f.BoolVar(&f.CodeInArg, "c", false, "take first argument as code to execute")
	f.BoolVar(&f.CompileOnly, "compileonly", false, "Parse/Compile but do not execute")
	f.BoolVar(&f.NoRc, "norc", false, "run elvish without invoking rc.elv")

	f.BoolVar(&f.Web, "web", false, "run backend of web interface")
	f.IntVar(&f.Port, "port", defaultWebPort, "the port of the web backend")

	f.BoolVar(&f.Daemon, "daemon", false, "run daemon instead of shell")

	f.StringVar(&f.Bin, "bin", "", "path to the elvish binary")
	f.StringVar(&f.DB, "db", "", "path to the database")
	f.StringVar(&f.Sock, "sock", "", "path to the daemon socket")

	return &f
}

// usage prints usage to the specified output. It modifies the flagSet; there is
// no API for getting the current output of a flag.FlagSet, so we can neither
// use the current output of f to output our own usage string, nor restore the
// previous value of f's output.
func usage(out io.Writer, f *flagSet) {
	f.SetOutput(out)
	fmt.Fprintln(out, "Usage: elvish [flags] [script]")
	fmt.Fprintln(out, "Supported flags:")
	f.PrintDefaults()
}

func Main(allArgs []string) int {
	flag := newFlagSet()
	err := flag.Parse(allArgs[1:])
	if err != nil {
		// Error and usage messages are already shown.
		return 2
	}

	// Handle flags common to all subprograms.

	if flag.CPUProfile != "" {
		f, err := os.Create(flag.CPUProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.Log != "" {
		err = util.SetOutputFile(flag.Log)
	} else if flag.LogPrefix != "" {
		err = util.SetOutputFile(flag.LogPrefix + strconv.Itoa(os.Getpid()))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return FindProgram(flag).Main(flag.Args())
}

// Program represents a subprogram.
type Program interface {
	// Main calls the subprogram with arguments. The return value will be used
	// as the exit status of the entire program.
	Main(args []string) int
}

// FindProgram finds a suitable Program according to flags. It does not have any
// side effects.
func FindProgram(flag *flagSet) Program {
	switch {
	case flag.Help:
		return ShowHelp{flag}
	case flag.Version:
		return ShowVersion{}
	case flag.BuildInfo:
		return ShowBuildInfo{flag.JSON}
	case flag.Daemon:
		if len(flag.Args()) > 0 {
			return ShowCorrectUsage{"arguments are not allowed with -daemon", flag}
		}
		return Daemon{inner: &daemon.Daemon{
			BinPath:       flag.Bin,
			DbPath:        flag.DB,
			SockPath:      flag.Sock,
			LogPathPrefix: flag.LogPrefix,
		}}
	case flag.Web:
		if len(flag.Args()) > 0 {
			return ShowCorrectUsage{"arguments are not allowed with -web", flag}
		}
		if flag.CodeInArg {
			return ShowCorrectUsage{"-c cannot be used together with -web", flag}
		}
		return web.New(flag.Bin, flag.Sock, flag.DB, flag.Port)
	default:
		return shell.New(flag.Bin, flag.Sock, flag.DB, flag.CodeInArg, flag.CompileOnly, flag.NoRc)
	}
}

// Package prog provides the entry point to Elvish. Its subpackages correspond
// to subprograms of Elvish.
package prog

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"src.elv.sh/pkg/logutil"
)

// Default port on which the web interface runs. The number is chosen because it
// resembles "elvi".
const defaultWebPort = 3171

// DeprecationLevel is a global flag that controls which deprecations to show.
// If its value is X, Elvish shows deprecations that should be shown for version
// 0.X.
var DeprecationLevel = 16

// Flags keeps command-line flags.
type Flags struct {
	Log, CPUProfile string

	Help, Version, BuildInfo, JSON bool

	CodeInArg, CompileOnly, NoRc bool
	RC                           string

	Web  bool
	Port int

	Daemon bool
	Forked int

	DB, Sock string
}

func newFlagSet(f *Flags) *flag.FlagSet {
	fs := flag.NewFlagSet("elvish", flag.ContinueOnError)
	// Error and usage will be printed explicitly.
	fs.SetOutput(io.Discard)

	fs.StringVar(&f.Log, "log", "", "a file to write debug log to except for the daemon")
	fs.StringVar(&f.CPUProfile, "cpuprofile", "", "write cpu profile to file")

	fs.BoolVar(&f.Help, "help", false, "show usage help and quit")
	fs.BoolVar(&f.Version, "version", false, "show version and quit")
	fs.BoolVar(&f.BuildInfo, "buildinfo", false, "show build info and quit")
	fs.BoolVar(&f.JSON, "json", false, "show output in JSON. Useful with -buildinfo.")

	// The `-i` option is for compatibility with POSIX shells so that programs, such as the `script`
	// command, will work when asked to launch an interactive Elvish shell.
	fs.Bool("i", false, "force interactive mode; currently ignored")
	fs.BoolVar(&f.CodeInArg, "c", false, "take first argument as code to execute")
	fs.BoolVar(&f.CompileOnly, "compileonly", false, "Parse/Compile but do not execute")
	fs.BoolVar(&f.NoRc, "norc", false, "run elvish without invoking rc.elv")
	fs.StringVar(&f.RC, "rc", "", "path to rc.elv")

	fs.BoolVar(&f.Web, "web", false, "run backend of web interface")
	fs.IntVar(&f.Port, "port", defaultWebPort, "the port of the web backend")

	fs.BoolVar(&f.Daemon, "daemon", false, "[internal flag] run the storage daemon instead of shell")

	fs.StringVar(&f.DB, "db", "", "[internal flag] path to the database")
	fs.StringVar(&f.Sock, "sock", "", "[internal flag] path to the daemon socket")

	fs.IntVar(&DeprecationLevel, "deprecation-level", DeprecationLevel, "show warnings for all features deprecated as of version 0.X")

	return fs
}

func usage(out io.Writer, f *flag.FlagSet) {
	fmt.Fprintln(out, "Usage: elvish [flags] [script]")
	fmt.Fprintln(out, "Supported flags:")
	f.SetOutput(out)
	f.PrintDefaults()
}

// Run parses command-line flags and runs the first applicable subprogram. It
// returns the exit status of the program.
func Run(fds [3]*os.File, args []string, programs ...Program) int {
	f := &Flags{}
	fs := newFlagSet(f)
	err := fs.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			// (*flag.FlagSet).Parse returns ErrHelp when -h or -help was
			// requested but *not* defined. Elvish defines -help, but not -h; so
			// this means that -h has been requested. Handle this by printing
			// the same message as an undefined flag.
			fmt.Fprintln(fds[2], "flag provided but not defined: -h")
		} else {
			fmt.Fprintln(fds[2], err)
		}
		usage(fds[2], fs)
		return 2
	}

	// Handle flags common to all subprograms.
	if f.CPUProfile != "" {
		f, err := os.Create(f.CPUProfile)
		if err != nil {
			fmt.Fprintln(fds[2], "Warning: cannot create CPU profile:", err)
			fmt.Fprintln(fds[2], "Continuing without CPU profiling.")
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	if f.Daemon {
		// We expect our stdout file handle is open on a unique log file for the daemon to write its
		// log messages. See daemon.Spawn() in pkg/daemon.
		logutil.SetOutput(fds[1])
	} else if f.Log != "" {
		err = logutil.SetOutputFile(f.Log)
		if err != nil {
			fmt.Fprintln(fds[2], err)
		}
	}

	if f.Help {
		usage(fds[1], fs)
		return 0
	}

	p := findProgram(f, programs)
	if p == nil {
		fmt.Fprintln(fds[2], "program bug: no suitable subprogram")
		return 2
	}

	err = p.Run(fds, f, fs.Args())
	if err == nil {
		return 0
	}
	if msg := err.Error(); msg != "" {
		fmt.Fprintln(fds[2], msg)
	}
	switch err := err.(type) {
	case badUsageError:
		usage(fds[2], fs)
	case exitError:
		return err.exit
	}
	return 2
}

func findProgram(f *Flags, programs []Program) Program {
	for _, program := range programs {
		if program.ShouldRun(f) {
			return program
		}
	}
	return nil
}

// BadUsage returns an error that may be returned by Program.Main, which
// requests the main program to print out a message, the usage information and
// exit with 2.
func BadUsage(msg string) error { return badUsageError{msg} }

type badUsageError struct{ msg string }

func (e badUsageError) Error() string { return e.msg }

// Exit returns an error that may be returned by Program.Main, which requests the
// main program to exit with the given code. If the exit code is 0, it returns nil.
func Exit(exit int) error {
	if exit == 0 {
		return nil
	}
	return exitError{exit}
}

type exitError struct{ exit int }

func (e exitError) Error() string { return "" }

// Program represents a subprogram.
type Program interface {
	// ShouldRun returns whether the subprogram should run.
	ShouldRun(f *Flags) bool
	// Run runs the subprogram.
	Run(fds [3]*os.File, f *Flags, args []string) error
}

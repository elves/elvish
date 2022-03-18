// Package prog provides the entry point to Elvish. Its subpackages correspond
// to subprograms of Elvish.
package prog

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"src.elv.sh/pkg/logutil"
)

// DeprecationLevel is a global flag that controls which deprecations to show.
// If its value is X, Elvish shows deprecations that should be shown for version
// 0.X.
var DeprecationLevel = 18

// Program represents a subprogram.
type Program interface {
	RegisterFlags(fs *FlagSet)
	// Run runs the subprogram.
	Run(fds [3]*os.File, args []string) error
}

func usage(out io.Writer, fs *flag.FlagSet) {
	fmt.Fprintln(out, "Usage: elvish [flags] [script] [args]")
	fmt.Fprintln(out, "Supported flags:")
	fs.SetOutput(out)
	fs.PrintDefaults()
}

// Run parses command-line flags and runs the first applicable subprogram. It
// returns the exit status of the program.
func Run(fds [3]*os.File, args []string, p Program) int {
	fs := flag.NewFlagSet("elvish", flag.ContinueOnError)
	// Error and usage will be printed explicitly.
	fs.SetOutput(io.Discard)

	var log string
	var help bool
	fs.StringVar(&log, "log", "",
		"Path to a file to write debug logs")
	fs.BoolVar(&help, "help", false,
		"Show usage help and quit")
	fs.IntVar(&DeprecationLevel, "deprecation-level", DeprecationLevel,
		"Show warnings for all features deprecated as of version 0.X")

	p.RegisterFlags(&FlagSet{FlagSet: fs})

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

	if log != "" {
		err = logutil.SetOutputFile(log)
		if err != nil {
			fmt.Fprintln(fds[2], err)
		}
	}

	if help {
		usage(fds[1], fs)
		return 0
	}

	err = p.Run(fds, fs.Args())
	if err == nil {
		return 0
	}
	if err == ErrNextProgram {
		err = errNoSuitableSubprogram
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

// Composite returns a Program that tries each of the given programs,
// terminating at the first one that doesn't return NotSuitable().
func Composite(programs ...Program) Program {
	return composite(programs)
}

type composite []Program

func (cp composite) RegisterFlags(f *FlagSet) {
	for _, p := range cp {
		p.RegisterFlags(f)
	}
}

func (cp composite) Run(fds [3]*os.File, args []string) error {
	for _, p := range cp {
		err := p.Run(fds, args)
		if err != ErrNextProgram {
			return err
		}
	}
	// If we have reached here, all subprograms have returned ErrNextProgram
	return ErrNextProgram
}

var errNoSuitableSubprogram = errors.New("internal error: no suitable subprogram")

// ErrNextProgram is a special error that may be returned by Program.Run that
// is part of a Composite program, indicating that the next program should be
// tried.
var ErrNextProgram = errors.New("next program")

// BadUsage returns a special error that may be returned by Program.Run. It
// causes the main function to print out a message, the usage information and
// exit with 2.
func BadUsage(msg string) error { return badUsageError{msg} }

type badUsageError struct{ msg string }

func (e badUsageError) Error() string { return e.msg }

// Exit returns a special error that may be returned by Program.Run. It causes
// the main function to exit with the given code without printing any error
// messages. Exit(0) returns nil.
func Exit(exit int) error {
	if exit == 0 {
		return nil
	}
	return exitError{exit}
}

type exitError struct{ exit int }

func (e exitError) Error() string { return "" }

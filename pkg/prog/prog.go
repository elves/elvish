/*
Package prog supports building testable, composable programs.

The main abstraction of this package is the [Program] interface, which can be
combined using [Composite] and run with [Run].

# Testability

The easy way to write a Go program is as follows:

  - Write code in the main function of the main package.
  - Access process-level context via globals defined in the [os] package,
    like [os.Args], [os.Stdin], [os.Stdout] and [os.Stderr].
  - Call [os.Exit] if the program should exit with a non-zero status.
  - Declare flags as global variables, with functions like [flag.String].

Programs written this way are hard to test, since they rely on process-level
states that are hard or impossible to emulate in tests.

With this package, the way to write a Go program is becomes a matter of
creating a type that implements the [Program] interface:

  - Write the "main" function in the Run method.
  - Context is available as arguments.
  - Return an error constructed from [Exit] to exit with a non-zero status.
  - Declare flags as fields of the type, and register them in the
    RegisterFlags method.

The [Program] can be run using the [Run] function, which takes care of
parsing flags, calling the Run method, and converting the error return value
into an exit code. Since the [Run] function itself takes the standard files
and command-line arguments as its function arguments, these can be emulated
easily in tests.

# Composability

Another advantage of this approach is composability. Elvish contains multiple
independent subprograms. The default subprogram is the shell; but if you run
Elvish with "elvish -buildinfo" or "elvish -daemon", they will invoke the
buildinfo or daemon subprogram instead of the shell.

Subprograms can also do something in addition to, rather than in place of,
other subprograms. One example is profiling support, which declares its own
flags like -cpuprofile and runs some extra code before and after other
subprograms to start and stop profiling.

Using this package, all the different subprograms can be implemented
separately and then composed into one using [Composite].

Other than keeping the codebase cleaner, this also enables an easy way to
provide alternative main packages that include or exclude a certain
subprogram. For example, profiling support requires importing a lot of
additional packages from the standard library, which increases the binary
size. As a result, the profiling subprogram is not included in the default
main package [src.elv.sh/cmd/elvish], but it is included in the alternative
main package [src.elv.sh/cmd/withpprof/elvish]. Binaries built from the
former main package is meaningfully smaller than the latter.

# Elvish-specific flag handling

As general as the [Program] abstraction is, this package has a bit of
Elvish-specific flag handling code:

-	[Run] handles some global flags not specific to any subprogram, like -log.

-	[FlagSet] handles some flags shared by multiple subprograms, like -json.

It's possible to split such code in this package, but doing so seems to require
a bit too much indirection to justify for the Elvish codebase.
*/
package prog

import (
	"flag"
	"fmt"
	"io"
	"os"

	"src.elv.sh/pkg/logutil"
)

// DeprecationLevel is a global flag that controls which deprecations to show.
// If its value is X, Elvish shows deprecations that should be shown for version
// 0.X.
var DeprecationLevel = 21

// Program represents a subprogram.
//
// This is the main abstraction provided by this package. See the package-level
// godoc for details.
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

// Run parses command-line flags and runs the [Program], returning the exit
// status. It also handles global flags that are not specific to any subprogram.
//
// It is supposed to be used from main functions like this:
//
//	func main() {
//		program := ...
//		os.Exit(prog.Run([3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args, program))
//	}
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
		if err == nil {
			defer logutil.SetOutput(io.Discard)
		} else {
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

// Composite returns a [Program] made up from subprograms. It starts from the
// first, continuing to the next as long as the subprogram returns an error
// created with [NextProgram].
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
	var cleanups []func([3]*os.File)
	for _, p := range cp {
		err := p.Run(fds, args)
		if np, ok := err.(nextProgramError); ok {
			cleanups = append(cleanups, np.cleanups...)
		} else {
			for i := len(cleanups) - 1; i >= 0; i-- {
				cleanups[i](fds)
			}
			return err
		}
	}
	// If we have reached here, all subprograms have returned ErrNextProgram
	return NextProgram(cleanups...)
}

// NextProgram returns a special error that may be returned by the Run method of
// a [Program] that is part of a [Composite] program, indicating that the next
// program should be tried. It can carry a list of cleanup functions that should
// be run in reverse order before the composite program finishes.
func NextProgram(cleanups ...func([3]*os.File)) error { return nextProgramError{cleanups} }

type nextProgramError struct{ cleanups []func([3]*os.File) }

// If this error ever gets printed, it has been bubbled to [Run] when all
// programs have returned this error type.
func (e nextProgramError) Error() string {
	return "internal error: no suitable subprogram"
}

// BadUsage returns a special error that may be returned by a [Program]'s Run
// method. It causes the main function to print out a message, the usage
// information and exit with 2.
func BadUsage(msg string) error { return badUsageError{msg} }

type badUsageError struct{ msg string }

func (e badUsageError) Error() string { return e.msg }

// Exit returns a special error that may be returned by a [Program]'s Run
// method. It causes the main function to exit with the given code without
// printing any error messages. Exit(0) returns nil.
func Exit(exit int) error {
	if exit == 0 {
		return nil
	}
	return exitError{exit}
}

type exitError struct{ exit int }

func (e exitError) Error() string { return "" }

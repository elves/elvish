package eval

import (
	"os"
	"os/exec"
	"strconv"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/prog"
)

// Command and process control.

// TODO(xiaq): Document "fg".

func init() {
	addBuiltinFns(map[string]any{
		// Command resolution
		"external":        external,
		"has-external":    hasExternal,
		"search-external": searchExternal,

		// Process control
		"fg":   fg,
		"exec": execFn,
		"exit": exit,
	})
}

//elvdoc:fn external
//
// ```elvish
// external $program
// ```
//
// Construct a callable value for the external program `$program`. Example:
//
// ```elvish-transcript
// ~> var x = (external man)
// ~> $x ls # opens the manpage for ls
// ```
//
// @cf has-external search-external

func external(cmd string) Callable {
	return NewExternalCmd(cmd)
}

//elvdoc:fn has-external
//
// ```elvish
// has-external $command
// ```
//
// Test whether `$command` names a valid external command. Examples (your output
// might differ):
//
// ```elvish-transcript
// ~> has-external cat
// ▶ $true
// ~> has-external lalala
// ▶ $false
// ```
//
// @cf external search-external

func hasExternal(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

//elvdoc:fn search-external
//
// ```elvish
// search-external $command
// ```
//
// Output the full path of the external `$command`. Throws an exception when not
// found. Example (your output might vary):
//
// ```elvish-transcript
// ~> search-external cat
// ▶ /bin/cat
// ```
//
// @cf external has-external

func searchExternal(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

//elvdoc:fn exit
//
// ```elvish
// exit $status?
// ```
//
// Exit the Elvish process with `$status` (defaulting to 0). The status must be
// in the range 0 to 255 inclusive.

// Can be overridden in tests.
var osExit = os.Exit

func exit(fm *Frame, codes ...int) error {
	code := 0
	switch len(codes) {
	case 0:
	case 1:
		code = codes[0]
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: len(codes)}
	}

	// On Windows the exit status has the range [0..2^32). Nonetheless, we
	// enforce the limits of traditional UNIX systems to avoid aliasing problems
	// on those systems (e.g., with respect to reporting signals as a reason for
	// the exit). Also, a wider range on Windows isn't particularly useful.
	if code < 0 || code > 255 {
		return errs.OutOfRange{What: "exit code", ValidLow: "0", ValidHigh: "255",
			Actual: strconv.Itoa(code)}
	}

	// We panic rather than directly calling os.Exit() because we want to run
	// any defer'ed functions. Such as those that capture profiling data. See
	// handlePanic() for where we perform the actual exit.
	fm.Evaler.PreExit()
	panic(prog.ExitStatus{Status: code})
}

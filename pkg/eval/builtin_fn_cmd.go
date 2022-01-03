package eval

import (
	"os"
	"os/exec"

	"src.elv.sh/pkg/eval/errs"
)

// Command and process control.

// TODO(xiaq): Document "fg".

func init() {
	addBuiltinFns(map[string]interface{}{
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
// Exit the Elvish process with `$status` (defaulting to 0).

func exit(fm *Frame, codes ...int) error {
	code := 0
	switch len(codes) {
	case 0:
	case 1:
		code = codes[0]
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: len(codes)}
	}

	preExit(fm)
	os.Exit(code)
	// Does not return
	panic("os.Exit returned")
}

func preExit(fm *Frame) {
	for _, hook := range fm.Evaler.BeforeExit {
		hook()
	}
}

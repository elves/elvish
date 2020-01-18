package eval

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Command and process control.

//elvdoc:fn external
//
// ```elvish
// external $program
// ```
//
// Construct a callable value for the external program `$program`. Example:
//
// ```elvish-transcript
// ~> x = (external man)
// ~> $x ls # opens the manpage for ls
// ```
//
// @cf has-external search-external

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

// TODO(xiaq): Document "fg".

//elvdoc:fn exec
//
// ```elvish
// exec $command?
// ```
//
// Replace the Elvish process with an external `$command`, defaulting to `elvish`.

//elvdoc:fn exit
//
// ```elvish
// exit $status?
// ```
//
// Exit the Elvish process with `$status` (defaulting to 0).

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

func external(cmd string) ExternalCmd {
	return ExternalCmd{cmd}
}

func hasExternal(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func searchExternal(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

func exit(fm *Frame, codes ...int) error {
	code := 0
	switch len(codes) {
	case 0:
	case 1:
		code = codes[0]
	default:
		return ErrArgs
	}

	preExit(fm)
	os.Exit(code)
	// Does not return
	panic("os.Exit returned")
}

func preExit(fm *Frame) {
	err := fm.DaemonClient.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

var errNotSupportedOnWindows = errors.New("not supported on Windows")

func notSupportedOnWindows() error {
	return errNotSupportedOnWindows
}

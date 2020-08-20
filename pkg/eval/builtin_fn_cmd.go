package eval

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
// ~> x = (external man)
// ~> $x ls # opens the manpage for ls
// ```
//
// @cf has-external search-external

func external(cmd string) ExternalCmd {
	return ExternalCmd{cmd}
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
// search-external &all=false $command
// ```
//
// Output the full path of the external `$command`. When `&all` is used it
// will output the full path for each dir in `$E:PATH` that contains the
// command. Throws an exception when not found. Example (your output might
// vary):
//
// ```elvish-transcript
// ~> search-external cat
// ▶ /bin/cat
// ~> search-external &all cat
// ▶ /bin/cat
// ▶ /usr/local/bin/cat
// ~> search-external &all no-such-cmd
// Exception: no-such-cmd: executable file not found in $E:PATH
// ```
//
// @cf external has-external

type searchOpts struct{ All bool }

func (opts *searchOpts) SetDefaultOptions() {}

// searchExternal implements the "search-external" command. See also
// EachExternal which this mimics when passed the `&all` option.
//
// TODO: Windows support (see fileIsExecutable).
func searchExternal(fm *Frame, opts searchOpts, cmd string) error {
	out := fm.OutputChan()

	// Note: Even if &all is used we must use exec.LookPath if the command
	// contains a path separator for consistency with how path resolution is
	// done when actually executing an external command.
	if !opts.All || strings.ContainsRune(cmd, '/') || strings.ContainsRune(cmd, '\\') {
		path, err := exec.LookPath(cmd)
		if err != nil {
			return err
		}
		out <- path
		return nil
	}

	// Assume we won't find a single dir containing the command.
	errReturn := fmt.Errorf("%s: executable file not found in $E:PATH", cmd)
	for _, dir := range searchPaths() {
		path := filepath.Join(dir, cmd)
		info, err := os.Stat(path)
		if err == nil && fileIsExecutable(info) {
			out <- path
			errReturn = nil // good news -- we found at least one instance in $E:PATH
		}
	}
	return errReturn
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
		return ErrArgs
	}

	preExit(fm)
	os.Exit(code)
	// Does not return
	panic("os.Exit returned")
}

func preExit(fm *Frame) {
	if fm.DaemonClient != nil {
		err := fm.DaemonClient.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

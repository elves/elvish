package eval

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/elves/elvish/eval/types"
)

// Command and process control.

var ErrNotInSameGroup = errors.New("not in the same process group")

func init() {
	addToBuiltinFns([]*BuiltinFn{
		// Command resolution
		{"resolve", resolveFn},
		{"has-external", hasExternal},
		{"search-external", searchExternal},

		// Process control
		{"fg", fg},
		{"exec", execFn},
		{"exit", exit},
	})
}

func resolveFn(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var cmd types.String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- resolve(string(cmd), ec)
}

func hasExternal(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var cmd types.String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	_, err := exec.LookPath(string(cmd))
	ec.OutputChan() <- types.Bool(err == nil)
}

func searchExternal(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var cmd types.String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	path, err := exec.LookPath(string(cmd))
	maybeThrow(err)

	out := ec.ports[1].Chan
	out <- types.String(path)
}

func exit(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var codes []int
	ScanArgsVariadic(args, &codes)
	TakeNoOpt(opts)

	doexit := func(i int) {
		preExit(ec)
		os.Exit(i)
	}
	switch len(codes) {
	case 0:
		doexit(0)
	case 1:
		doexit(codes[0])
	default:
		throw(ErrArgs)
	}
}

func preExit(ec *Frame) {
	err := ec.DaemonClient.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

var errNotSupportedOnWindows = errors.New("not supported on Windows")

func notSupportedOnWindows(ec *Frame, args []types.Value, opts map[string]types.Value) {
	throw(errNotSupportedOnWindows)
}

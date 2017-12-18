package eval

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon/api"
)

var ErrDaemonOffline = errors.New("daemon is offline")

func makeDaemonNamespace(daemon *api.Client) Namespace {
	// Obtain process ID
	daemonPid := func() Value {
		pid, err := daemon.Pid()
		maybeThrow(err)
		return String(strconv.Itoa(pid))
	}

	return Namespace{
		"pid":  MakeRoVariableFromCallback(daemonPid),
		"sock": NewRoVariable(String(daemon.SockPath())),

		"spawn" + FnSuffix: NewRoVariable(&BuiltinFn{"daemon:spawn", daemonSpawn}),
	}
}

func daemonSpawn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)
	maybeThrow(ec.ToSpawn.Spawn())
}

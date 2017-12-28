package eval

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon"
)

// ErrDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var ErrDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

func makeDaemonNs(daemon *daemon.Client) Ns {
	// Obtain process ID
	daemonPid := func() Value {
		pid, err := daemon.Pid()
		maybeThrow(err)
		return String(strconv.Itoa(pid))
	}

	return Ns{
		"pid":  MakeRoVariableFromCallback(daemonPid),
		"sock": NewRoVariable(String(daemon.SockPath())),

		"spawn" + FnSuffix: NewRoVariable(&BuiltinFn{"daemon:spawn", daemonSpawn}),
	}
}

func daemonSpawn(ec *Frame, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)
	if ec.DaemonSpawner == nil {
		throw(ErrDontKnowHowToSpawnDaemon)
	}
	maybeThrow(ec.DaemonSpawner.Spawn())
}

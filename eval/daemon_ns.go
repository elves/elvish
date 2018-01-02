package eval

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/types"
)

// ErrDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var ErrDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

func makeDaemonNs(daemon *daemon.Client) Ns {
	// Obtain process ID
	daemonPid := func() types.Value {
		pid, err := daemon.Pid()
		maybeThrow(err)
		return types.String(strconv.Itoa(pid))
	}

	return Ns{
		"pid":  MakeRoVariableFromCallback(daemonPid),
		"sock": NewRoVariable(types.String(daemon.SockPath())),

		"spawn" + FnSuffix: NewRoVariable(&BuiltinFn{"daemon:spawn", daemonSpawn}),
	}
}

func daemonSpawn(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)
	if ec.DaemonSpawner == nil {
		throw(ErrDontKnowHowToSpawnDaemon)
	}
	maybeThrow(ec.DaemonSpawner.Spawn())
}

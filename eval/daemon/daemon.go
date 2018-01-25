// Package daemon implements the builtin daemon: module.
package daemon

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	daemonp "github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/util"
)

// errDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var errDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

// Ns makes the daemon: namespace.
func Ns(daemon *daemon.Client, spawner *daemonp.Daemon) eval.Ns {
	// Obtain process ID
	daemonPid := func() types.Value {
		pid, err := daemon.Pid()
		if err != nil {
			util.Throw(err)
		}
		return string(strconv.Itoa(pid))
	}

	daemonSpawn := func(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
		eval.TakeNoArg(args)
		eval.TakeNoOpt(opts)
		if spawner == nil {
			util.Throw(errDontKnowHowToSpawnDaemon)
		}
		err := spawner.Spawn()
		if err != nil {
			util.Throw(err)
		}
	}
	return eval.Ns{
		"pid":  vartypes.NewRoCallback(daemonPid),
		"sock": vartypes.NewRo(string(daemon.SockPath())),

		"spawn" + eval.FnSuffix: vartypes.NewRo(&eval.BuiltinFn{"daemon:spawn", daemonSpawn}),
	}
}

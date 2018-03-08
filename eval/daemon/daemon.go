// Package daemon implements the builtin daemon: module.
package daemon

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	daemonp "github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/util"
)

// errDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var errDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

// Ns makes the daemon: namespace.
func Ns(daemon *daemon.Client, spawner *daemonp.Daemon) eval.Ns {
	// Obtain process ID
	getPid := func() interface{} {
		pid, err := daemon.Pid()
		if err != nil {
			util.Throw(err)
		}
		return string(strconv.Itoa(pid))
	}

	spawn := func() error {
		if spawner == nil {
			return errDontKnowHowToSpawnDaemon
		}
		return spawner.Spawn()
	}

	return eval.Ns{
		"pid":  vars.FromGet(getPid),
		"sock": vars.NewRo(string(daemon.SockPath())),
	}.AddBuiltinFn("daemon:", "spawn", spawn)
}

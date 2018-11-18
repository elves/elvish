// Package daemon implements the builtin daemon: module.
package daemon

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	daemonp "github.com/elves/elvish/program/daemon"
)

// errDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var errDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

// Ns makes the daemon: namespace.
func Ns(daemon *daemon.Client, spawner *daemonp.Daemon) eval.Ns {
	getPid := func() (string, error) {
		pid, err := daemon.Pid()
		return string(strconv.Itoa(pid)), err
	}

	spawn := func() error {
		if spawner == nil {
			return errDontKnowHowToSpawnDaemon
		}
		return spawner.Spawn()
	}

	// TODO: Deprecate the variable in favor of the function.
	getPidVar := func() interface{} {
		pid, err := getPid()
		if err != nil {
			return "-1"
		}
		return pid
	}

	return eval.Ns{
		"pid":  vars.FromGet(getPidVar),
		"sock": vars.NewReadOnly(string(daemon.SockPath())),
	}.AddBuiltinFns("daemon:", map[string]interface{}{
		"pid":   getPid,
		"spawn": spawn,
	})
}

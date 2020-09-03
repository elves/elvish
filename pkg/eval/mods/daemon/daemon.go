// Package daemon implements the builtin daemon: module.
package daemon

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vars"
)

// errDontKnowHowToSpawnDaemon is thrown by daemon:spawn when the Evaler's
// DaemonSpawner field is nil.
var errDontKnowHowToSpawnDaemon = errors.New("don't know how to spawn daemon")

// Ns makes the daemon: namespace.
func Ns(d daemon.Client, spawnCfg *daemon.SpawnConfig) eval.Ns {
	getPid := func() (string, error) {
		pid, err := d.Pid()
		return string(strconv.Itoa(pid)), err
	}

	spawn := func() error {
		if spawnCfg == nil {
			return errDontKnowHowToSpawnDaemon
		}
		return daemon.Spawn(spawnCfg)
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
		"sock": vars.NewReadOnly(string(d.SockPath())),
	}.AddGoFns("daemon:", map[string]interface{}{
		"pid":   getPid,
		"spawn": spawn,
	})
}

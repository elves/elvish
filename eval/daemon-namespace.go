package eval

import (
	"errors"
	"net/rpc"
	"strconv"

	"github.com/elves/elvish/daemon/api"
)

var ErrDaemonOffline = errors.New("daemon is offline")

func makeDaemonNamespace(daemon *rpc.Client) Namespace {
	// Obtain process ID
	daemonPid := func() Value {
		if daemon == nil {
			throw(ErrDaemonOffline)
		}
		req := &api.PidRequest{}
		res := &api.PidResponse{}
		err := daemon.Call(api.ServiceName+".Pid", req, res)
		maybeThrow(err)
		return String(strconv.Itoa(res.Pid))
	}

	return Namespace{
		"pid": MakeRoVariableFromCallback(daemonPid),
	}
}

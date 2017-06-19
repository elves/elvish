package eval

import (
	"net/rpc"
	"strconv"

	"github.com/elves/elvish/daemon/api"
)

func makeDaemonNamespace(daemon *rpc.Client) Namespace {
	// Obtain process ID
	req := &api.PidRequest{}
	res := &api.PidResponse{}
	daemon.Call(api.ServiceName+".Pid", req, res)
	pid := res.Pid

	return Namespace{
		"pid": NewRoVariable(String(strconv.Itoa(pid))),
	}
}

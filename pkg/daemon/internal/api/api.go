// Package api defines types and constants useful for the API between the daemon
// service and client.
package api

import (
	"src.elv.sh/pkg/store/storedefs"
)

// Version is the API version. It should be bumped any time the API changes.
const Version = -93

// ServiceName is the name of the RPC service exposed by the daemon.
const ServiceName = "Daemon"

// Basic requests.

type VersionRequest struct{}

type VersionResponse struct {
	Version int
}

type PidRequest struct{}

type PidResponse struct {
	Pid int
}

// Cmd requests.

type NextCmdSeqRequest struct{}

type NextCmdSeqResponse struct {
	Seq int
}

type AddCmdRequest struct {
	Text string
}

type AddCmdResponse struct {
	Seq int
}

type DelCmdRequest struct {
	Seq int
}

type DelCmdResponse struct {
}

type CmdRequest struct {
	Seq int
}

type CmdResponse struct {
	Text string
}

type CmdsRequest struct {
	From int
	Upto int
}

type CmdsResponse struct {
	Cmds []string
}

type CmdsWithSeqRequest struct {
	From int
	Upto int
}

type CmdsWithSeqResponse struct {
	Cmds []storedefs.Cmd
}

type NextCmdRequest struct {
	From   int
	Prefix string
}

type NextCmdResponse struct {
	Seq  int
	Text string
}

type PrevCmdRequest struct {
	Upto   int
	Prefix string
}

type PrevCmdResponse struct {
	Seq  int
	Text string
}

// Dir requests.

type AddDirRequest struct {
	Dir       string
	IncFactor float64
}

type AddDirResponse struct{}

type DelDirRequest struct {
	Dir string
}

type DelDirResponse struct{}

type DirsRequest struct {
	Blacklist map[string]struct{}
}

type DirsResponse struct {
	Dirs []storedefs.Dir
}

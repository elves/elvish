// Package API provides the API to the daemon RPC service.
package api

const (
	// ServiceName is the name of the RPC service exposed by the daemon.
	ServiceName = "Daemon"

	// Version is the API version. It should be bumped any time the API changes.
	Version = -99
)

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

/*
// Dir requests.

type AddDir struct {
	Dir       string
	IncFactor float64
}

type GetDirs struct {
	Blacklist map[string]struct{}
}

// SharedVar requests.

type GetSharedVar struct {
	Name string
}

type SetSharedVar struct {
	Name  string
	Value string
}

type DelSharedVar struct {
	Name string
}

type Request struct {
	GetPid *GetPid

	NextCmdSeq  *NextCmdSeq
	AddCmd      *AddCmd
	GetCmds     *GetCmds
	GetFirstCmd *GetFirstCmd
	GetLastCmd  *GetLastCmd

	AddDir  *AddDir
	GetDirs *GetDirs

	GetSharedVar *GetSharedVar
	SetSharedVar *SetSharedVar
	DelSharedVar *DelSharedVar
}
*/

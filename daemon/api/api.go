// Package API provides the API to the daemon RPC service.
package api

import "github.com/elves/elvish/store"

const (
	// ServiceName is the name of the RPC service exposed by the daemon.
	ServiceName = "Daemon"

	// Version is the API version. It should be bumped any time the API changes.
	Version = -97
)

// Basic requests.

type VersionRequest struct{}

type VersionResponse struct {
	Version int
}

func (c *Client) Version() (int, error) {
	req := &VersionRequest{}
	res := &VersionResponse{}
	err := c.CallDaemon("Version", req, res)
	return res.Version, err
}

type PidRequest struct{}

type PidResponse struct {
	Pid int
}

func (c *Client) Pid() (int, error) {
	req := &PidRequest{}
	res := &PidResponse{}
	err := c.CallDaemon("Pid", req, res)
	return res.Pid, err
}

// Cmd requests.

type NextCmdSeqRequest struct{}

type NextCmdSeqResponse struct {
	Seq int
}

func (c *Client) NextCmdSeq() (int, error) {
	req := &NextCmdRequest{}
	res := &NextCmdSeqResponse{}
	err := c.CallDaemon("NextCmdSeq", req, res)
	return res.Seq, err
}

type AddCmdRequest struct {
	Text string
}

type AddCmdResponse struct {
	Seq int
}

func (c *Client) AddCmd(text string) (int, error) {
	req := &AddCmdRequest{text}
	res := &AddCmdResponse{}
	err := c.CallDaemon("AddCmd", req, res)
	return res.Seq, err
}

type CmdRequest struct {
	Seq int
}

type CmdResponse struct {
	Text string
}

func (c *Client) Cmd(seq int) (string, error) {
	req := &CmdRequest{seq}
	res := &CmdResponse{}
	err := c.CallDaemon("Cmd", req, res)
	return res.Text, err
}

type CmdsRequest struct {
	From int
	Upto int
}

type CmdsResponse struct {
	Cmds []string
}

func (c *Client) Cmds(from, upto int) ([]string, error) {
	req := &CmdsRequest{from, upto}
	res := &CmdsResponse{}
	err := c.CallDaemon("Cmds", req, res)
	return res.Cmds, err
}

type NextCmdRequest struct {
	From   int
	Prefix string
}

type NextCmdResponse struct {
	Seq  int
	Text string
}

func (c *Client) NextCmd(from int, prefix string) (int, string, error) {
	req := &NextCmdRequest{from, prefix}
	res := &NextCmdResponse{}
	err := c.CallDaemon("NextCmd", req, res)
	return res.Seq, res.Text, err
}

type PrevCmdRequest struct {
	Upto   int
	Prefix string
}

type PrevCmdResponse struct {
	Seq  int
	Text string
}

func (c *Client) PrevCmd(upto int, prefix string) (int, string, error) {
	req := &PrevCmdRequest{upto, prefix}
	res := &PrevCmdResponse{}
	err := c.CallDaemon("PrevCmd", req, res)
	return res.Seq, res.Text, err
}

// Dir requests.

type AddDirRequest struct {
	Dir       string
	IncFactor float64
}

type AddDirResponse struct{}

func (c *Client) AddDir(dir string, incFactor float64) error {
	req := &AddDirRequest{dir, incFactor}
	res := &AddDirResponse{}
	err := c.CallDaemon("AddDir", req, res)
	return err
}

type DirsRequest struct {
	Blacklist map[string]struct{}
}

type DirsResponse struct {
	Dirs []store.Dir
}

func (c *Client) Dirs(blacklist map[string]struct{}) ([]store.Dir, error) {
	req := &DirsRequest{blacklist}
	res := &DirsResponse{}
	err := c.CallDaemon("Dirs", req, res)
	return res.Dirs, err
}

// SharedVar requests.

type SharedVarRequest struct {
	Name string
}

type SharedVarResponse struct {
	Value string
}

func (c *Client) SharedVar(name string) (string, error) {
	req := &SharedVarRequest{name}
	res := &SharedVarResponse{}
	err := c.CallDaemon("SharedVar", req, res)
	return res.Value, err
}

type SetSharedVarRequest struct {
	Name  string
	Value string
}

type SetSharedVarResponse struct{}

func (c *Client) SetSharedVar(name, value string) error {
	req := &SetSharedVarRequest{name, value}
	res := &SetSharedVarResponse{}
	return c.CallDaemon("SetSharedVar", req, res)
}

type DelSharedVarRequest struct {
	Name string
}

type DelSharedVarResponse struct{}

func (c *Client) DelSharedVar(name string) error {
	req := &DelSharedVarRequest{}
	res := &DelSharedVarResponse{}
	return c.CallDaemon("DelSharedVar", req, res)
}

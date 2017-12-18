package api

import (
	"errors"
	"net/rpc"
	"sync"

	"github.com/elves/elvish/store/storedefs"
)

var ErrDaemonOffline = errors.New("daemon offline")

type Client struct {
	sockPath  string
	rpcClient *rpc.Client
	waits     sync.WaitGroup
}

func NewClient(sockPath string) *Client {
	return &Client{sockPath, nil, sync.WaitGroup{}}
}

func (c *Client) SockPath() string {
	return c.sockPath
}

func (c *Client) Waits() *sync.WaitGroup {
	return &c.waits
}

func (c *Client) Close() error {
	c.waits.Wait()
	rc := c.rpcClient
	c.rpcClient = nil
	if rc != nil {
		return rc.Close()
	}
	return nil
}

func (c *Client) call(f string, req, res interface{}) error {
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.rpcClient.Call(ServiceName+"."+f, req, res)
	if err == rpc.ErrShutdown {
		// Clear rpcClient so as to reconnect next time
		c.rpcClient = nil
	}
	return err
}

func (c *Client) connect() error {
	if c.rpcClient == nil {
		conn, err := dial(c.sockPath)
		if err != nil {
			return err
		}
		c.rpcClient = rpc.NewClient(conn)
	}
	return nil
}

// Convenience methods for RPC methods.

func (c *Client) Version() (int, error) {
	req := &VersionRequest{}
	res := &VersionResponse{}
	err := c.call("Version", req, res)
	return res.Version, err
}

func (c *Client) Pid() (int, error) {
	req := &PidRequest{}
	res := &PidResponse{}
	err := c.call("Pid", req, res)
	return res.Pid, err
}

func (c *Client) NextCmdSeq() (int, error) {
	req := &NextCmdRequest{}
	res := &NextCmdSeqResponse{}
	err := c.call("NextCmdSeq", req, res)
	return res.Seq, err
}

func (c *Client) AddCmd(text string) (int, error) {
	req := &AddCmdRequest{text}
	res := &AddCmdResponse{}
	err := c.call("AddCmd", req, res)
	return res.Seq, err
}

func (c *Client) Cmd(seq int) (string, error) {
	req := &CmdRequest{seq}
	res := &CmdResponse{}
	err := c.call("Cmd", req, res)
	return res.Text, err
}

func (c *Client) Cmds(from, upto int) ([]string, error) {
	req := &CmdsRequest{from, upto}
	res := &CmdsResponse{}
	err := c.call("Cmds", req, res)
	return res.Cmds, err
}

func (c *Client) NextCmd(from int, prefix string) (int, string, error) {
	req := &NextCmdRequest{from, prefix}
	res := &NextCmdResponse{}
	err := c.call("NextCmd", req, res)
	return res.Seq, res.Text, err
}

func (c *Client) PrevCmd(upto int, prefix string) (int, string, error) {
	req := &PrevCmdRequest{upto, prefix}
	res := &PrevCmdResponse{}
	err := c.call("PrevCmd", req, res)
	return res.Seq, res.Text, err
}

func (c *Client) AddDir(dir string, incFactor float64) error {
	req := &AddDirRequest{dir, incFactor}
	res := &AddDirResponse{}
	err := c.call("AddDir", req, res)
	return err
}

func (c *Client) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	req := &DirsRequest{blacklist}
	res := &DirsResponse{}
	err := c.call("Dirs", req, res)
	return res.Dirs, err
}

func (c *Client) SharedVar(name string) (string, error) {
	req := &SharedVarRequest{name}
	res := &SharedVarResponse{}
	err := c.call("SharedVar", req, res)
	return res.Value, err
}

func (c *Client) SetSharedVar(name, value string) error {
	req := &SetSharedVarRequest{name, value}
	res := &SetSharedVarResponse{}
	return c.call("SetSharedVar", req, res)
}

func (c *Client) DelSharedVar(name string) error {
	req := &DelSharedVarRequest{}
	res := &DelSharedVarResponse{}
	return c.call("DelSharedVar", req, res)
}

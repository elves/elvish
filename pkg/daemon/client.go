package daemon

import (
	"errors"
	"net"
	"sync"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/rpc"
	"src.elv.sh/pkg/store/storedefs"
)

const retriesOnShutdown = 3

var (
	// ErrDaemonUnreachable is returned when the daemon cannot be reached after
	// several retries.
	ErrDaemonUnreachable = errors.New("daemon offline")
)

// Implementation of the Client interface.
type client struct {
	sockPath  string
	rpcClient *rpc.Client
	waits     sync.WaitGroup
}

// NewClient creates a new Client instance that talks to the socket. Connection
// creation is deferred to the first request.
func NewClient(sockPath string) daemondefs.Client {
	return &client{sockPath, nil, sync.WaitGroup{}}
}

// SockPath returns the socket path that the Client talks to. If the client is
// nil, it returns an empty string.
func (c *client) SockPath() string {
	return c.sockPath
}

// ResetConn resets the current connection. A new connection will be established
// the next time a request is made. If the client is nil, it does nothing.
func (c *client) ResetConn() error {
	if c.rpcClient == nil {
		return nil
	}
	rc := c.rpcClient
	c.rpcClient = nil
	return rc.Close()
}

// Close waits for all outstanding requests to finish and close the connection.
// If the client is nil, it does nothing and returns nil.
func (c *client) Close() error {
	c.waits.Wait()
	return c.ResetConn()
}

func (c *client) call(f string, req, res any) error {
	c.waits.Add(1)
	defer c.waits.Done()

	for attempt := 0; attempt < retriesOnShutdown; attempt++ {
		if c.rpcClient == nil {
			conn, err := net.Dial("unix", c.sockPath)
			if err != nil {
				return err
			}
			c.rpcClient = rpc.NewClient(conn)
		}

		err := c.rpcClient.Call(api.ServiceName+"."+f, req, res)
		if err == rpc.ErrShutdown {
			// Clear rpcClient so as to reconnect next time
			c.rpcClient = nil
			continue
		} else {
			return err
		}
	}
	return ErrDaemonUnreachable
}

// Convenience methods for RPC methods. These are quite repetitive; when the
// number of RPC calls grow above some threshold, a code generator should be
// written to generate them.

func (c *client) Version() (int, error) {
	req := &api.VersionRequest{}
	res := &api.VersionResponse{}
	err := c.call("Version", req, res)
	return res.Version, err
}

func (c *client) Pid() (int, error) {
	req := &api.PidRequest{}
	res := &api.PidResponse{}
	err := c.call("Pid", req, res)
	return res.Pid, err
}

func (c *client) NextCmdSeq() (int, error) {
	req := &api.NextCmdRequest{}
	res := &api.NextCmdSeqResponse{}
	err := c.call("NextCmdSeq", req, res)
	return res.Seq, err
}

func (c *client) AddCmd(text string) (int, error) {
	req := &api.AddCmdRequest{Text: text}
	res := &api.AddCmdResponse{}
	err := c.call("AddCmd", req, res)
	return res.Seq, err
}

func (c *client) DelCmd(seq int) error {
	req := &api.DelCmdRequest{Seq: seq}
	res := &api.DelCmdResponse{}
	err := c.call("DelCmd", req, res)
	return err
}

func (c *client) Cmd(seq int) (string, error) {
	req := &api.CmdRequest{Seq: seq}
	res := &api.CmdResponse{}
	err := c.call("Cmd", req, res)
	return res.Text, err
}

func (c *client) CmdsWithSeq(from, upto int) ([]storedefs.Cmd, error) {
	req := &api.CmdsWithSeqRequest{From: from, Upto: upto}
	res := &api.CmdsWithSeqResponse{}
	err := c.call("CmdsWithSeq", req, res)
	return res.Cmds, err
}

func (c *client) NextCmd(from int, prefix string) (storedefs.Cmd, error) {
	req := &api.NextCmdRequest{From: from, Prefix: prefix}
	res := &api.NextCmdResponse{}
	err := c.call("NextCmd", req, res)
	return storedefs.Cmd{Text: res.Text, Seq: res.Seq}, err
}

func (c *client) PrevCmd(upto int, prefix string) (storedefs.Cmd, error) {
	req := &api.PrevCmdRequest{Upto: upto, Prefix: prefix}
	res := &api.PrevCmdResponse{}
	err := c.call("PrevCmd", req, res)
	return storedefs.Cmd{Text: res.Text, Seq: res.Seq}, err
}

func (c *client) AddDir(dir string, incFactor float64) error {
	req := &api.AddDirRequest{Dir: dir, IncFactor: incFactor}
	res := &api.AddDirResponse{}
	err := c.call("AddDir", req, res)
	return err
}

func (c *client) DelDir(dir string) error {
	req := &api.DelDirRequest{Dir: dir}
	res := &api.DelDirResponse{}
	err := c.call("DelDir", req, res)
	return err
}

func (c *client) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	req := &api.DirsRequest{Blacklist: blacklist}
	res := &api.DirsResponse{}
	err := c.call("Dirs", req, res)
	return res.Dirs, err
}

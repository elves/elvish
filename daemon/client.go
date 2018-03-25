package daemon

import (
	"errors"
	"net/rpc"
	"sync"

	"github.com/elves/elvish/store/storedefs"
)

const retriesOnShutdown = 3

var (
	// ErrClientNotInitialized is returned when the Client is not initialized.
	ErrClientNotInitialized = errors.New("client not initialized")
	// ErrDaemonUnreachable is returned when the daemon cannot be reached after
	// several retries.
	ErrDaemonUnreachable = errors.New("daemon offline")
)

// Client is a client to the Elvish daemon. A nil *Client is safe to use.
type Client struct {
	sockPath  string
	rpcClient *rpc.Client
	waits     sync.WaitGroup
}

var _ storedefs.Store = (*Client)(nil)

// NewClient creates a new Client instance that talks to the socket. Connection
// creation is deferred to the first request.
func NewClient(sockPath string) *Client {
	return &Client{sockPath, nil, sync.WaitGroup{}}
}

// SockPath returns the socket path that the Client talks to. If the client is
// nil, it returns an empty string.
func (c *Client) SockPath() string {
	if c == nil {
		return ""
	}
	return c.sockPath
}

// ResetConn resets the current connection. A new connection will be established
// the next time a request is made. If the client is nil, it does nothing.
func (c *Client) ResetConn() error {
	if c == nil || c.rpcClient == nil {
		return nil
	}
	rc := c.rpcClient
	c.rpcClient = nil
	return rc.Close()
}

// Close waits for all outstanding requests to finish and close the connection.
// If the client is nil, it does nothing and returns nil.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.waits.Wait()
	return c.ResetConn()
}

func (c *Client) call(f string, req, res interface{}) error {
	if c == nil {
		return ErrClientNotInitialized
	}
	c.waits.Add(1)
	defer c.waits.Done()

	for attempt := 0; attempt < retriesOnShutdown; attempt++ {
		if c.rpcClient == nil {
			conn, err := dial(c.sockPath)
			if err != nil {
				return err
			}
			c.rpcClient = rpc.NewClient(conn)
		}

		err := c.rpcClient.Call(ServiceName+"."+f, req, res)
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

func (c *Client) DelCmd(seq int) error {
	req := &DelCmdRequest{seq}
	res := &DelCmdResponse{}
	err := c.call("DelCmd", req, res)
	return err
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

func (c *Client) DelDir(dir string) error {
	req := &DelDirRequest{dir}
	res := &DelDirResponse{}
	err := c.call("DelDir", req, res)
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

// +build elv_daemon_stub

package daemon

import (
	"errors"

	"src.elv.sh/pkg/store"
)

var (
	// This symbol exists so the sole use of this error in pkg/shell/runtime.detectDaemon() doesn't
	// have to also import the rpc package.
	ErrShutdown   = errors.New("should never be used")
	ErrNoCommands = errors.New("no commands")
)

// Client represents a daemon client.
type Client interface {
	store.Store

	ResetConn() error
	Close() error

	Pid() (int, error)
	SockPath() string
	Version() (int, error)
}

// Implementation of the Client interface.
type client struct{}

// NewClient creates a new Client instance that talks to the socket. Connection
// creation is deferred to the first request.
func NewClient(sockPath string) Client {
	return &client{}
}

// SockPath returns the socket path that the Client talks to. If the client is
// nil, it returns an empty string.
func (c *client) SockPath() string {
	return ""
}

// ResetConn resets the current connection. A new connection will be established
// the next time a request is made. If the client is nil, it does nothing.
func (c *client) ResetConn() error {
	return nil
}

// Close waits for all outstanding requests to finish and close the connection.
// If the client is nil, it does nothing and returns nil.
func (c *client) Close() error {
	return nil
}

// Convenience methods for RPC methods. These are quite repetitive; when the
// number of RPC calls grow above some threshold, a code generator should be
// written to generate them.

func (c *client) Version() (int, error) {
	return 0, nil
}

func (c *client) Pid() (int, error) {
	return 0, nil
}

func (c *client) NextCmdSeq() (int, error) {
	return 0, nil
}

func (c *client) AddCmd(text string) (int, error) {
	return 0, nil
}

func (c *client) DelCmd(seq int) error {
	return nil
}

func (c *client) Cmd(seq int) (string, error) {
	return "", nil
}

func (c *client) CmdsWithSeq(from, upto int) ([]store.Cmd, error) {
	return []store.Cmd{}, ErrNoCommands
}

func (c *client) NextCmd(from int, prefix string) (store.Cmd, error) {
	return store.Cmd{}, ErrNoCommands
}

func (c *client) PrevCmd(upto int, prefix string) (store.Cmd, error) {
	return store.Cmd{}, ErrNoCommands
}

func (c *client) AddDir(dir string, incFactor float64) error {
	return nil
}

func (c *client) DelDir(dir string) error {
	return nil
}

func (c *client) Dirs(blacklist map[string]struct{}) ([]store.Dir, error) {
	return []store.Dir{}, nil
}

func (c *client) SharedVar(name string) (string, error) {
	return "", nil
}

func (c *client) SetSharedVar(name, value string) error {
	return nil
}

func (c *client) DelSharedVar(name string) error {
	return nil
}

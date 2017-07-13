package api

import (
	"errors"
	"net/rpc"
	"sync"
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

func (c *Client) CallDaemon(f string, req, res interface{}) error {
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

func (c *Client) Close() error {
	c.waits.Wait()
	rc := c.rpcClient
	c.rpcClient = nil
	if rc != nil {
		return rc.Close()
	}
	return nil
}

func (c *Client) connect() error {
	if c.rpcClient == nil {
		rpcClient, err := rpc.Dial("unix", c.sockPath)
		if err != nil {
			return err
		}
		c.rpcClient = rpcClient
	}
	return nil
}

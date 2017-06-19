package api

import (
	"errors"
	"net/rpc"
)

var ErrDaemonOffline = errors.New("daemon offline")

type Client struct {
	sockPath  string
	rpcClient *rpc.Client
}

func NewClient(sockPath string) *Client {
	return &Client{sockPath, nil}
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
	return c.rpcClient.Close()
}

func (c *Client) connect() error {
	rpcClient, err := rpc.Dial("unix", c.sockPath)
	if err != nil {
		return err
	}
	c.rpcClient = rpcClient
	return nil
}

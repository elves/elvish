// Package service implements server and client for the elvishd service.
package service

import (
	"errors"
	"net"
	"net/rpc"
)

const (
	Version = "0"
)

var (
	VersionMismatch = errors.New("version mismatch")
)

type Elvishd struct{}

// Serve starts the RPC server on listener. Serve blocks.
func Serve(listener net.Listener) {
	server := &Elvishd{}
	rpc.Register(server)
	rpc.Accept(listener)
}

// Version replies with a string that identify the RPC version.
func (e *Elvishd) Version(arg struct{}, reply *string) error {
	*reply = Version
	return nil
}

// Echo replies with an exact duplicate of the argument string. Could be useful
// for testing.
func (e *Elvishd) Echo(arg string, reply *string) error {
	*reply = arg
	return nil
}

// Client wraps rpc.Client with type-safe wrappers.
type Client struct {
	rc *rpc.Client
}

// Dial establishes RPC connection and check for version mismatch.
func Dial(network, address string) (Client, error) {
	rc, err := rpc.Dial(network, address)
	if err != nil {
		return Client{}, err
	}
	c := Client{rc}
	var version string
	c.Version(struct{}{}, &version)
	if version != Version {
		return Client{}, VersionMismatch
	}
	return c, nil
}

func (c Client) Version(arg struct{}, reply *string) error {
	return c.rc.Call("Elvishd.Version", arg, reply)
}

func (c Client) Echo(arg string, reply *string) error {
	return c.rc.Call("Elvishd.Echo", arg, reply)
}

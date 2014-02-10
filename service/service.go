// Package service implements server and client for the elvishd service.
package service

import (
	"errors"
	"net"
	"net/rpc"

	"github.com/coopernurse/gorp"
)

const (
	Version = "0"
)

var (
	VersionMismatch = errors.New("version mismatch")
	UniVarNotFound  = errors.New("universal variable not found")
)

type Elvishd struct {
	dbmap *gorp.DbMap
}

type UniVar struct {
	Name  string
	Value string // TODO(xiaq): support arbitrary elvish value
}

// Serve starts the RPC server on listener. Serve blocks.
func Serve(listener net.Listener, dbmap *gorp.DbMap) error {
	dbmap.AddTable(UniVar{}).SetKeys(false, "Name")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}

	server := &Elvishd{dbmap}
	rpc.Register(server)
	rpc.Accept(listener)
	return nil
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

// GetUniVar replies with the value of the universal variable with the name
// arg. If the named variable does not exist or there is a database error, an
// error is returned instead.
func (e *Elvishd) GetUniVar(arg string, reply *string) error {
	univar, err := e.dbmap.Get(UniVar{}, arg)
	if err != nil {
		return err
	}
	if univar == nil {
		return UniVarNotFound
	}
	*reply = univar.(*UniVar).Value
	return nil
}

// SetUniVar sets the universal variable to the given value. It is created if
// nonexistent. If there is a database error, an error is returned.
func (e *Elvishd) SetUniVar(arg *UniVar, reply *struct{}) error {
	current, err := e.dbmap.Get(UniVar{}, arg.Name)
	if err != nil {
		return err
	}
	if current == nil {
		return e.dbmap.Insert(arg)
	} else {
		_, err = e.dbmap.Update(arg)
		return err
	}
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

func (c Client) GetUniVar(arg string, reply *string) error {
	return c.rc.Call("Elvishd.GetUniVar", arg, reply)
}

func (c Client) SetUniVar(arg *UniVar, reply *struct{}) error {
	return c.rc.Call("Elvishd.SetUniVar", arg, reply)
}

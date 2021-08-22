//go:build gccgo
// +build gccgo

package eval

import "errors"

var errPluginNotImplemented = errors.New("plugin not implemented")

type pluginStub struct{}

func pluginOpen(name string) (pluginStub, error) {
	return pluginStub{}, errPluginNotImplemented
}

func (pluginStub) Lookup(symName string) (interface{}, error) {
	return nil, errPluginNotImplemented
}

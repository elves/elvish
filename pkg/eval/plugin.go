//go:build !gccgo
// +build !gccgo

package eval

import "plugin"

var pluginOpen = plugin.Open

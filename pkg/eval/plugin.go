//go:build !gccgo

package eval

import "plugin"

var pluginOpen = plugin.Open

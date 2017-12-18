// +build !windows,!plan9

package daemon

import "net"

func listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}

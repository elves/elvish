// +build !windows,!plan9

package client

import "net"

func dial(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}

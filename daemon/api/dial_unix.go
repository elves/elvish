// +build !windows,!plan9

package api

import "net"

func dial(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}

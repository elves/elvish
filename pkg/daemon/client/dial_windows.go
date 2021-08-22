package client

import (
	"net"
	"os"
)

func dial(path string) (net.Conn, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return net.Dial("tcp", string(buf))
}

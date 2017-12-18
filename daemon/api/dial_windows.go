package api

import (
	"io/ioutil"
	"net"
)

func dial(path string) (net.Conn, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return net.Dial("tcp", string(buf))
}

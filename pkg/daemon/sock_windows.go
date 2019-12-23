package daemon

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func listen(path string) (net.Listener, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0006)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		file.Close()
		err2 := os.Remove(path)
		if err2 != nil {
			logger.Println("Failed to remove sock file after failure to listen", err2)
		}
		return nil, err
	}
	_, err = fmt.Fprint(file, listener.Addr())
	if err != nil {
		logger.Println("Failed to write to sock file after listening", err)
		listener.Close()
		file.Close()
		err2 := os.Remove(path)
		if err2 != nil {
			logger.Println("Failed to remove sock file after failure to listen", err2)
		}
		return nil, err
	}
	file.Close()
	return listener, nil
}

func dial(path string) (net.Conn, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return net.Dial("tcp", string(buf))
}

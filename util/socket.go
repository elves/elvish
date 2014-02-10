package util

import (
	"fmt"
	"os/user"
)

// SocketName returns the path of the per-user Unix socket elvish and elvishd
// use for communication.
func SocketName() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/tmp/elvishd-%s.sock", user.Username), nil
}

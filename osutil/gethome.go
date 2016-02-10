package osutil

import (
	"fmt"
	"os/user"
)

func GetHome(uname string) (string, error) {
	var u *user.User
	var err error
	if uname == "" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(uname)
	}
	if err != nil {
		return "", fmt.Errorf("can't resolve ~%s: %s", uname, err.Error())
	}
	return u.HomeDir, nil
}

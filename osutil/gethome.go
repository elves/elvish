package osutil

import (
	"fmt"
	"os"
	"os/user"
	"strings"
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
		if uname == "" {
			// Use $HOME as fallback
			home := os.Getenv("HOME")
			if home != "" {
				return strings.TrimRight(home, "/"), nil
			}
		}
		return "", fmt.Errorf("can't resolve ~%s: %s", uname, err.Error())
	}
	return strings.TrimRight(u.HomeDir, "/"), nil
}

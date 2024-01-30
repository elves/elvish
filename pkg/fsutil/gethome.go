package fsutil

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	"src.elv.sh/pkg/env"
)

// GetHome finds the home directory of a specified user. When given an empty
// string, it finds the home directory of the current user.
func GetHome(uname string) (string, error) {
	if uname == "" {
		// Use $HOME as override if we are looking for the home of the current
		// variable.
		home := os.Getenv(env.HOME)
		if home != "" {
			if runtime.GOOS == "windows" {
				return strings.TrimRight(home, "/\\"), nil
			} else {
				return strings.TrimRight(home, "/"), nil
			}
		}
	}

	// Look up the user.
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
	return strings.TrimRight(u.HomeDir, "/"), nil
}

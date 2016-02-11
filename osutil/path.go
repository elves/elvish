package osutil

import (
	"os"
	"strings"
)

// Getwd returns path of the working directory in a format suitable as the
// prompt.
func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	home, err := GetHome("")
	if err == nil {
		if pwd == home {
			return "~"
		} else if strings.HasPrefix(pwd, home+"/") {
			return "~" + pwd[len(home):]
		}
	}
	return pwd
}

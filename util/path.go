package util

import (
	"os"
	"strings"
)

func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	home := os.Getenv("HOME")
	home = strings.TrimRight(home, "/")
	if len(pwd) >= len(home) && pwd[:len(home)] == home {
		return "~" + pwd[len(home):]
	} else {
		return pwd
	}
}

package store

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrEmptyHOME = errors.New("Environment variable HOME is empty")
)

// ensureDataDir ensures Elvish's data directory exists, creating it if
// necessary. It returns the path to the data directory (never with a
// trailing slash) and possible error.
func EnsureDataDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", ErrEmptyHOME
	}
	home = strings.TrimRight(home, "/")
	ddir := home + "/.elvish"
	return ddir, os.MkdirAll(ddir, 0700)
}

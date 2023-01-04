//go:build windows || plan9 || js

package lscolors

import (
	"errors"
)

var errNotSupportedOnNonUnix = errors.New("not supported on non-Unix OS")

func createNamedPipe(fname string) error {
	return errNotSupportedOnNonUnix
}

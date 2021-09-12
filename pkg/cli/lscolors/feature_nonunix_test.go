//go:build windows || plan9 || js
// +build windows plan9 js

package lscolors

import (
	"errors"
)

var errNotSupportedOnNonUNIX = errors.New("not supported on non-UNIX OS")

func createNamedPipe(fname string) error {
	return errNotSupportedOnNonUNIX
}

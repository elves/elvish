package eval

import "errors"

var (
	execFn = notSupportedOnWindows
	fg     = notSupportedOnWindows
)

var errNotSupportedOnWindows = errors.New("not supported on Windows")

func notSupportedOnWindows() error {
	return errNotSupportedOnWindows
}

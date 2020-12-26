package eval

import "errors"

var errNotSupportedOnWindows = errors.New("not supported on Windows")

func execFn(...interface{}) error {
	return errNotSupportedOnWindows
}

func fg(...int) error {
	return errNotSupportedOnWindows
}

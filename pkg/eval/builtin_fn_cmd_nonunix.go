//go:build !unix

package eval

import "errors"

var errOnlySupportedOnUNIX = errors.New("only supported on UNIX")

func execFn(...any) error {
	return errOnlySupportedOnUNIX
}

func fg(...int) error {
	return errOnlySupportedOnUNIX
}

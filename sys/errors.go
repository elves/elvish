package sys

import "errors"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidAction  = errors.New("invalid action")
)

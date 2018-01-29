package vartypes

import (
	"errors"
)

var errMustBeString = errors.New("must be string")

type stringVar struct {
	ptr *string
}

// NewString creates a variable from a string pointer. The Variable can only be
// set to a String value, and modifications are reflected in the passed string.
func NewString(ps *string) Variable {
	return stringVar{ps}
}

func (sv stringVar) Get() interface{} {
	return string(*sv.ptr)
}

func (sv stringVar) Set(v interface{}) error {
	if s, ok := v.(string); ok {
		*sv.ptr = string(s)
		return nil
	}
	return errMustBeString
}

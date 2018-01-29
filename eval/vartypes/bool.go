package vartypes

import (
	"errors"
)

var errMustBeBool = errors.New("must be bool")

type boolVar struct {
	ptr *bool
}

// NewBool returns a variable backed by a *bool. The Set method of the variable
// only accept bool arguments.
func NewBool(ptr *bool) Variable {
	return boolVar{ptr}
}

func (bv boolVar) Get() interface{} {
	return *bv.ptr
}

func (bv boolVar) Set(v interface{}) error {
	if b, ok := v.(bool); ok {
		*bv.ptr = b
		return nil
	}
	return errMustBeBool
}

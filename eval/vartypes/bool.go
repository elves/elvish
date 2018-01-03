package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var errMustBeBool = errors.New("must be bool")

type boolVar struct {
	ptr *bool
}

func NewBool(ptr *bool) Variable {
	return boolVar{ptr}
}

func (bv boolVar) Get() types.Value {
	return types.Bool(*bv.ptr)
}

func (bv boolVar) Set(v types.Value) error {
	if b, ok := v.(types.Bool); ok {
		*bv.ptr = bool(b)
		return nil
	}
	return errMustBeBool
}

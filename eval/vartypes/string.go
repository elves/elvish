package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
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

func (sv stringVar) Get() types.Value {
	return string(*sv.ptr)
}

func (sv stringVar) Set(v types.Value) error {
	if s, ok := v.(string); ok {
		*sv.ptr = string(s)
		return nil
	}
	return errMustBeString
}

package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var errRoCannotBeSet = errors.New("read-only variable; cannot be set")

type roVariable struct {
	value types.Value
}

func NewRoVariable(v types.Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val types.Value) error {
	return errRoCannotBeSet
}

func (rv roVariable) Get() types.Value {
	return rv.value
}

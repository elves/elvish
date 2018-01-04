package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var errRoCannotBeSet = errors.New("read-only variable; cannot be set")

type ro struct {
	value types.Value
}

func NewRo(v types.Value) Variable {
	return ro{v}
}

func (rv ro) Set(val types.Value) error {
	return errRoCannotBeSet
}

func (rv ro) Get() types.Value {
	return rv.value
}

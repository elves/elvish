package vars

import (
	"errors"
)

var errRoCannotBeSet = errors.New("read-only variable; cannot be set")

type ro struct {
	value interface{}
}

func NewRo(v interface{}) Var {
	return ro{v}
}

func (rv ro) Set(val interface{}) error {
	return errRoCannotBeSet
}

func (rv ro) Get() interface{} {
	return rv.value
}

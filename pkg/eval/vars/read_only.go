package vars

import (
	"errors"
)

// ErrSetReadOnlyVar is returned by the Set method of a read-only variable.
var ErrSetReadOnlyVar = errors.New("read-only variable; cannot be set")

type readOnly struct {
	value interface{}
}

// NewReadOnly creates a variable that is read-only and always returns an error
// on Set.
func NewReadOnly(v interface{}) Var {
	return readOnly{v}
}

func (rv readOnly) Set(val interface{}) error {
	return ErrSetReadOnlyVar
}

func (rv readOnly) Get() interface{} {
	return rv.value
}

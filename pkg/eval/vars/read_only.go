package vars

import (
	"fmt"
)

// ErrSetReadOnlyVar is returned by the Set method of a read-only variable.
type ErrSetReadOnlyVar struct {
	VarName string
}

func (e *ErrSetReadOnlyVar) Error() string {
	return fmt.Sprintf("%s: read-only variable cannot be set", e.VarName)
}

type readOnly struct {
	name  string
	value interface{}
}

// NewReadOnly creates a variable that is read-only and always returns an error
// on Set.
func NewReadOnly(name string, value interface{}) Var {
	return readOnly{name, value}
}

func (rv readOnly) Set(val interface{}) error {
	return &ErrSetReadOnlyVar{rv.name}
}

func (rv readOnly) Get() interface{} {
	return rv.value
}

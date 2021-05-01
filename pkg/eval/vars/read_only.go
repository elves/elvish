package vars

import (
	"src.elv.sh/pkg/eval/errs"
)

type readOnly struct {
	value interface{}
}

// NewReadOnly creates a variable that is read-only and always returns an error
// on Set.
func NewReadOnly(v interface{}) Var {
	return readOnly{v}
}

func (rv readOnly) Set(val interface{}) error {
	return errs.SetReadOnlyVar{}
}

func (rv readOnly) Get() interface{} {
	return rv.value
}

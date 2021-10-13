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

// IsReadOnly returns whether v is a read-only variable.
func IsReadOnly(v Var) bool {
	switch v.(type) {
	case readOnly:
		return true
	case roCallback:
		return true
	default:
		return false
	}
}

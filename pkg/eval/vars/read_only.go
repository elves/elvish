package vars

import (
	"src.elv.sh/pkg/eval/errs"
)

type readOnly struct {
	value any
}

// NewReadOnly creates a variable that is read-only and always returns an error
// on Set.
func NewReadOnly(v any) Var {
	return readOnly{v}
}

func (rv readOnly) Set(val any) error {
	return errs.SetReadOnlyVar{}
}

func (rv readOnly) Get() any {
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

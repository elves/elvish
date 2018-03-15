package vals

import (
	"errors"
	"fmt"
)

// Concatter wraps the Concat method. See Concat for how it is used.
type Concatter interface {
	// Concat concatenates the receiver with another value, the receiver being
	// the left operand. If concatenation is not supported for the given value,
	// the method can return the special error type ErrCatNotImplemented.
	Concat(v interface{}) (interface{}, error)
}

// RConcatter wraps the RConcat method. See Concat for how it is used.
type RConcatter interface {
	RConcat(v interface{}) (interface{}, error)
}

// ErrConcatNotImplemented is a special error value used to signal that
// concatenation is not implemented. See Concat for how it is used.
var ErrConcatNotImplemented = errors.New("cat not implemented")

// Concat concatenates two values. If both operands are strings, it returns lhs
// + rhs, nil. If the left operand implements Concatter, it calls
// lhs.Concat(rhs). If lhs doesn't implement the interface or returned
// ErrConcatNotImplemented, it then calls rhs.RConcat(lhs). If all attempts
// fail, it returns nil and an error.
func Concat(lhs, rhs interface{}) (interface{}, error) {
	if v, ok := tryConcatBuiltins(lhs, rhs); ok {
		return v, nil
	}

	if lhs, ok := lhs.(Concatter); ok {
		v, err := lhs.Concat(rhs)
		if err != ErrConcatNotImplemented {
			return v, err
		}
	}

	if rhs, ok := rhs.(RConcatter); ok {
		v, err := rhs.RConcat(lhs)
		if err != ErrConcatNotImplemented {
			return v, err
		}
	}

	return nil, fmt.Errorf("unsupported concat: %s and %s", Kind(lhs), Kind(rhs))
}

func tryConcatBuiltins(lhs, rhs interface{}) (interface{}, bool) {
	switch lhs := lhs.(type) {
	case string:
		if rhs, ok := rhs.(string); ok {
			return lhs + rhs, true
		}
	}

	return nil, false
}

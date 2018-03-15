package vals

import (
	"errors"
	"fmt"
)

type (
	Concatter interface {
		Concat(v interface{}) (interface{}, error)
	}

	RConcatter interface {
		RConcat(v interface{}) (interface{}, error)
	}
)

var ErrCatNotImplemented = errors.New("cat not implemented")

func Concat(lhs, rhs interface{}) (interface{}, error) {
	if v, ok := tryConcatBuiltins(lhs, rhs); ok {
		return v, nil
	}

	if lhs, ok := lhs.(Concatter); ok {
		v, err := lhs.Concat(rhs)
		if err != ErrCatNotImplemented {
			return v, err
		}
	}

	if rhs, ok := rhs.(RConcatter); ok {
		v, err := rhs.RConcat(lhs)
		if err != ErrCatNotImplemented {
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

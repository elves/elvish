package types

import "errors"

// Iterator wraps the Iterate method.
type Iterator interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v Value) bool)
}

var errCannotIterate = errors.New("cannot be iterated")

func Iterate(v Value, f func(Value) bool) error {
	switch v := v.(type) {
	case Iterator:
		v.Iterate(f)
		return nil
	}
	return errCannotIterate
}

func Collect(it Value) ([]Value, error) {
	var vs []Value
	if len := Len(it); len >= 0 {
		vs = make([]Value, 0, len)
	}
	err := Iterate(it, func(v Value) bool {
		vs = append(vs, v)
		return true
	})
	return vs, err
}

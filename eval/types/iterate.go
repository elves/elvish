package types

import "errors"

// Iterator wraps the Iterate method.
type Iterator interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v Value) bool)
}

func Iterate(v Value, f func(Value) bool) error {
	switch v := v.(type) {
	case string:
		for _, r := range v {
			b := f(string(r))
			if !b {
				break
			}
		}
		return nil
	case Iterator:
		v.Iterate(f)
		return nil
	default:
		return errors.New(Kind(v) + " cannot be iterated")
	}
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

package types

import "errors"

// Assocer wraps the Assoc method.
type Assocer interface {
	// Assoc returns a slightly modified version of the receiver with key k
	// associated with value v.
	Assoc(k, v Value) (Value, error)
}

var errAssocUnsupported = errors.New("assoc is not supported")

func Assoc(a, k, v Value) (Value, error) {
	switch a := a.(type) {
	case string:
	case Assocer:
		return a.Assoc(k, v)
	}
	return nil, errAssocUnsupported
}

var ErrReplacementMustBeString = errors.New("replacement must be string")

func assocString(s string, k, v Value) (Value, error) {
	i, j, err := indexString(s, k)
	if err != nil {
		return nil, err
	}
	repl, ok := v.(string)
	if !ok {
		return nil, ErrReplacementMustBeString
	}
	return s[:i] + repl + s[j:], nil
}

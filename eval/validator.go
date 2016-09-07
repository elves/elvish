package eval

import "errors"

var mustBeListOfFnValue = errors.New("must be a list of fn")

func IsListOfFnValue(v Value) error {
	li, ok := v.(ListLike)
	if !ok {
		return mustBeListOfFnValue
	}
	listok := true
	li.Iterate(func(v Value) bool {
		if _, ok := v.(FnValue); !ok {
			listok = false
			return false
		}
		return true
	})
	if !listok {
		return mustBeListOfFnValue
	}
	return nil
}

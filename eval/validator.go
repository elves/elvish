package eval

import "errors"

var (
	mustBeListOfFnValue = errors.New("must be a list of fn")
	errShouldBeFn       = errors.New("should be function")
	errShouldBeBool     = errors.New("should be bool")
)

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

func ShouldBeFn(v Value) error {
	if _, ok := v.(Fn); !ok {
		return errShouldBeFn
	}
	return nil
}

func ShouldBeBool(v Value) error {
	if _, ok := v.(Bool); !ok {
		return errShouldBeBool
	}
	return nil
}

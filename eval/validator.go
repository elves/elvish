package eval

import "errors"

var (
	errShouldBeList = errors.New("should be list")
	errShouldBeMap  = errors.New("should be map")
	errShouldBeFn   = errors.New("should be function")
	errShouldBeBool = errors.New("should be bool")
)

func ShouldBeList(v Value) error {
	if _, ok := v.(List); !ok {
		return errShouldBeList
	}
	return nil
}

func ShouldBeMap(v Value) error {
	if _, ok := v.(Map); !ok {
		return errShouldBeMap
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

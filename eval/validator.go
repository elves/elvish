package eval

import (
	"errors"
)

var (
	errShouldBeFn = errors.New("should be function")
	errShouldBeNs = errors.New("should be ns")
)

func ShouldBeFn(v interface{}) error {
	if _, ok := v.(Callable); !ok {
		return errShouldBeFn
	}
	return nil
}

func ShouldBeNs(v interface{}) error {
	if _, ok := v.(Ns); !ok {
		return errShouldBeNs
	}
	return nil
}

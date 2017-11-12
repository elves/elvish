package eval

import (
	"errors"
	"strconv"
)

var (
	errShouldBeList   = errors.New("should be list")
	errShouldBeMap    = errors.New("should be map")
	errShouldBeFn     = errors.New("should be function")
	errShouldBeBool   = errors.New("should be bool")
	errShouldBeNumber = errors.New("should be number")
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
	if _, ok := v.(Callable); !ok {
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

func ShouldBeNumber(v Value) error {
	if _, ok := v.(String); !ok {
		return errShouldBeNumber
	}
	_, err := strconv.ParseFloat(string(v.(String)), 64)
	return err
}

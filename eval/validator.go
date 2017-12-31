package eval

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/eval/types"
)

var (
	errShouldBeList   = errors.New("should be list")
	errShouldBeMap    = errors.New("should be map")
	errShouldBeFn     = errors.New("should be function")
	errShouldBeNs     = errors.New("should be ns")
	errShouldBeBool   = errors.New("should be bool")
	errShouldBeNumber = errors.New("should be number")
)

func ShouldBeList(v types.Value) error {
	if _, ok := v.(types.List); !ok {
		return errShouldBeList
	}
	return nil
}

func ShouldBeMap(v types.Value) error {
	if _, ok := v.(Map); !ok {
		return errShouldBeMap
	}
	return nil
}

func ShouldBeFn(v types.Value) error {
	if _, ok := v.(Callable); !ok {
		return errShouldBeFn
	}
	return nil
}

func ShouldBeNs(v types.Value) error {
	if _, ok := v.(Ns); !ok {
		return errShouldBeNs
	}
	return nil
}

func ShouldBeBool(v types.Value) error {
	if _, ok := v.(Bool); !ok {
		return errShouldBeBool
	}
	return nil
}

func ShouldBeNumber(v types.Value) error {
	if _, ok := v.(String); !ok {
		return errShouldBeNumber
	}
	_, err := strconv.ParseFloat(string(v.(String)), 64)
	return err
}

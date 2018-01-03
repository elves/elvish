package eval

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var (
	errShouldBeFn = errors.New("should be function")
	errShouldBeNs = errors.New("should be ns")
)

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

package vartypes

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/eval/types"
)

var (
	errShouldBeList   = errors.New("should be list")
	errShouldBeMap    = errors.New("should be map")
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
	if _, ok := v.(types.Map); !ok {
		return errShouldBeMap
	}
	return nil
}

func ShouldBeBool(v types.Value) error {
	if _, ok := v.(types.Bool); !ok {
		return errShouldBeBool
	}
	return nil
}

func ShouldBeNumber(v types.Value) error {
	if _, ok := v.(types.String); !ok {
		return errShouldBeNumber
	}
	_, err := strconv.ParseFloat(string(v.(types.String)), 64)
	return err
}

package vartypes

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

var (
	errShouldBeList   = errors.New("should be list")
	errShouldBeMap    = errors.New("should be map")
	errShouldBeBool   = errors.New("should be bool")
	errShouldBeNumber = errors.New("should be number")
)

func ShouldBeList(v types.Value) error {
	if _, ok := v.(vector.Vector); !ok {
		return errShouldBeList
	}
	return nil
}

func ShouldBeMap(v types.Value) error {
	if _, ok := v.(hashmap.Map); !ok {
		return errShouldBeMap
	}
	return nil
}

func ShouldBeBool(v types.Value) error {
	if _, ok := v.(bool); !ok {
		return errShouldBeBool
	}
	return nil
}

func ShouldBeNumber(v types.Value) error {
	if _, ok := v.(string); !ok {
		return errShouldBeNumber
	}
	_, err := strconv.ParseFloat(string(v.(string)), 64)
	return err
}

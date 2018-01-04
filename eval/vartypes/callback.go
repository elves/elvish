package vartypes

import "github.com/elves/elvish/eval/types"

type callback struct {
	set func(types.Value) error
	get func() types.Value
}

// NewCallback makes a variable from a set callback and a get callback.
func NewCallback(set func(types.Value) error, get func() types.Value) Variable {
	return &callback{set, get}
}

func (cv *callback) Set(val types.Value) error {
	return cv.set(val)
}

func (cv *callback) Get() types.Value {
	return cv.get()
}

type roCallback func() types.Value

// NewRoCallback makes a read-only variable from a get callback.
func NewRoCallback(get func() types.Value) Variable {
	return roCallback(get)
}

func (cv roCallback) Set(types.Value) error {
	return errRoCannotBeSet
}

func (cv roCallback) Get() types.Value {
	return cv()
}

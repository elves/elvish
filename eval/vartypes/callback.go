package vartypes

import "github.com/elves/elvish/eval/types"

type cbVariable struct {
	set func(types.Value) error
	get func() types.Value
}

// NewCallbackVariable makes a variable from a set callback and a get callback.
func NewCallbackVariable(set func(types.Value) error, get func() types.Value) Variable {
	return &cbVariable{set, get}
}

func (cv *cbVariable) Set(val types.Value) error {
	return cv.set(val)
}

func (cv *cbVariable) Get() types.Value {
	return cv.get()
}

type roCbVariable func() types.Value

// NewRoCallbackVariable makes a read-only variable from a get callback.
func NewRoCallbackVariable(get func() types.Value) Variable {
	return roCbVariable(get)
}

func (cv roCbVariable) Set(types.Value) error {
	return errRoCannotBeSet
}

func (cv roCbVariable) Get() types.Value {
	return cv()
}

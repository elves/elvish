package vars

import (
	"src.elv.sh/pkg/eval/errs"
)

type callback struct {
	set func(any) error
	get func() any
}

// FromSetGet makes a variable from a set callback and a get callback.
func FromSetGet(set func(any) error, get func() any) Var {
	return &callback{set, get}
}

func (cv *callback) Set(val any) error {
	return cv.set(val)
}

func (cv *callback) Get() any {
	return cv.get()
}

type roCallback func() any

// FromGet makes a variable from a get callback. The variable is read-only.
func FromGet(get func() any) Var {
	return roCallback(get)
}

func (cv roCallback) Set(any) error {
	return errs.SetReadOnlyVar{}
}

func (cv roCallback) Get() any {
	return cv()
}

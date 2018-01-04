package vartypes

import "github.com/elves/elvish/eval/types"

type ptr struct {
	valuePtr *types.Value
}

func (pv ptr) Set(val types.Value) error {
	*pv.valuePtr = val
	return nil
}

func (pv ptr) Get() types.Value {
	return *pv.valuePtr
}

func NewPtr(v types.Value) Variable {
	return ptr{&v}
}

type validatedPtr struct {
	valuePtr  *types.Value
	validator func(types.Value) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewValidatedPtr(v types.Value, vld func(types.Value) error) Variable {
	return validatedPtr{&v, vld}
}

func (iv validatedPtr) Set(val types.Value) error {
	if err := iv.validator(val); err != nil {
		return invalidValueError{err}
	}
	*iv.valuePtr = val
	return nil
}

func (iv validatedPtr) Get() types.Value {
	return *iv.valuePtr
}

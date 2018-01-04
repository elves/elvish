package vartypes

import "github.com/elves/elvish/eval/types"

type ptrVariable struct {
	valuePtr *types.Value
}

func (pv ptrVariable) Set(val types.Value) error {
	*pv.valuePtr = val
	return nil
}

func (pv ptrVariable) Get() types.Value {
	return *pv.valuePtr
}

func NewPtrVariable(v types.Value) Variable {
	return ptrVariable{&v}
}

type validatedPtrVariable struct {
	valuePtr  *types.Value
	validator func(types.Value) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewValidatedPtrVariable(v types.Value, vld func(types.Value) error) Variable {
	return validatedPtrVariable{&v, vld}
}

func (iv validatedPtrVariable) Set(val types.Value) error {
	if err := iv.validator(val); err != nil {
		return invalidValueError{err}
	}
	*iv.valuePtr = val
	return nil
}

func (iv validatedPtrVariable) Get() types.Value {
	return *iv.valuePtr
}

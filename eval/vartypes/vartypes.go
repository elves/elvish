// Package vartypes contains basic types for manipulating Elvish variables.
package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var errRoCannotBeSet = errors.New("read-only variable; cannot be set")

// Variable represents an Elvish variable.
type Variable interface {
	Set(v types.Value) error
	Get() types.Value
}

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

type roVariable struct {
	value types.Value
}

func NewRoVariable(v types.Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val types.Value) error {
	return errRoCannotBeSet
}

func (rv roVariable) Get() types.Value {
	return rv.value
}

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

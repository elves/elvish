package vartypes

import (
	"errors"

	"github.com/elves/elvish/eval/types"
)

var errRoCannotBeSet = errors.New("read-only variable; cannot be set")

// Variable represents an Elvish variable.
type Variable interface {
	Set(v types.Value)
	Get() types.Value
}

type ptrVariable struct {
	valuePtr  *types.Value
	validator func(types.Value) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewPtrVariable(v types.Value) Variable {
	return NewPtrVariableWithValidator(v, nil)
}

func NewPtrVariableWithValidator(v types.Value, vld func(types.Value) error) Variable {
	return ptrVariable{&v, vld}
}

func (iv ptrVariable) Set(val types.Value) {
	if iv.validator != nil {
		if err := iv.validator(val); err != nil {
			throw(invalidValueError{err})
		}
	}
	*iv.valuePtr = val
}

func (iv ptrVariable) Get() types.Value {
	return *iv.valuePtr
}

type roVariable struct {
	value types.Value
}

func NewRoVariable(v types.Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val types.Value) {
	throw(errRoCannotBeSet)
}

func (rv roVariable) Get() types.Value {
	return rv.value
}

type cbVariable struct {
	set func(types.Value)
	get func() types.Value
}

// MakeVariableFromCallback makes a variable from a set callback and a get
// callback.
func MakeVariableFromCallback(set func(types.Value), get func() types.Value) Variable {
	return &cbVariable{set, get}
}

func (cv *cbVariable) Set(val types.Value) {
	cv.set(val)
}

func (cv *cbVariable) Get() types.Value {
	return cv.get()
}

type roCbVariable func() types.Value

// MakeRoVariableFromCallback makes a read-only variable from a get callback.
func MakeRoVariableFromCallback(get func() types.Value) Variable {
	return roCbVariable(get)
}

func (cv roCbVariable) Set(types.Value) {
	throw(errRoCannotBeSet)
}

func (cv roCbVariable) Get() types.Value {
	return cv()
}

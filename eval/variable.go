package eval

import (
	"errors"
	"os"
)

var (
	ErrRoCannotBeSet = errors.New("read-only variable; cannot be set")
)

// Variable represents an elvish variable.
type Variable interface {
	Set(v Value)
	Get() Value
}

type ptrVariable struct {
	valuePtr  *Value
	validator func(Value) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewPtrVariable(v Value) Variable {
	return NewPtrVariableWithValidator(v, nil)
}

func NewPtrVariableWithValidator(v Value, vld func(Value) error) Variable {
	return ptrVariable{&v, vld}
}

func (iv ptrVariable) Set(val Value) {
	if iv.validator != nil {
		if err := iv.validator(val); err != nil {
			throw(invalidValueError{err})
		}
	}
	*iv.valuePtr = val
}

func (iv ptrVariable) Get() Value {
	return *iv.valuePtr
}

type roVariable struct {
	value Value
}

func NewRoVariable(v Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val Value) {
	throw(ErrRoCannotBeSet)
}

func (rv roVariable) Get() Value {
	return rv.value
}

type cbVariable struct {
	set func(Value)
	get func() Value
}

// MakeVariableFromCallback makes a variable from a set callback and a get
// callback.
func MakeVariableFromCallback(set func(Value), get func() Value) Variable {
	return &cbVariable{set, get}
}

func (cv *cbVariable) Set(val Value) {
	cv.set(val)
}

func (cv *cbVariable) Get() Value {
	return cv.get()
}

type roCbVariable func() Value

// MakeRoVariableFromCallback makes a read-only variable from a get callback.
func MakeRoVariableFromCallback(get func() Value) Variable {
	return roCbVariable(get)
}

func (cv roCbVariable) Set(Value) {
	throw(ErrRoCannotBeSet)
}

func (cv roCbVariable) Get() Value {
	return cv()
}

// elemVariable is an element of a IndexSetter.
type elemVariable struct {
	container IndexSetter
	index     Value
}

func (ev elemVariable) Set(val Value) {
	ev.container.IndexSet(ev.index, val)
}

func (ev elemVariable) Get() Value {
	return ev.container.IndexOne(ev.index)
}

type envVariable struct {
	name string
}

func (ev envVariable) Set(val Value) {
	os.Setenv(ev.name, ToString(val))
}

func (ev envVariable) Get() Value {
	return String(os.Getenv(ev.name))
}

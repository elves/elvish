package eval

import (
	"reflect"

	"github.com/elves/elvish/eval/vartypes"
)

type ptrVariable struct {
	ptr interface{}
}

// NewVariableFromPtr creates a variable from a pointer. The variable is kept in
// sync with the value the pointer points to, using elvToGo and goToElv
// conversions when Get and Set.
func NewVariableFromPtr(ptr interface{}) vartypes.Variable {
	return ptrVariable{ptr}
}

// Get returns the value pointed by the pointer, after conversion using goToElv.
func (v ptrVariable) Get() interface{} {
	return goToElv(reflect.Indirect(reflect.ValueOf(v.ptr)).Interface())
}

// Get sets the value pointed by the pointer, after conversion using elvToGo.
func (v ptrVariable) Set(val interface{}) error {
	return scanElvToGo(val, v.ptr)
}

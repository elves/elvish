package vars

import (
	"reflect"

	"github.com/elves/elvish/eval/vals"
)

type ptr struct {
	ptr interface{}
}

// FromPtr creates a variable from a pointer. The variable is kept in sync
// with the value the pointer points to, using elvToGo and goToElv conversions
// when Get and Set.
func FromPtr(p interface{}) Var {
	return ptr{p}
}

// Get returns the value pointed by the pointer, after conversion using FromGo.
func (v ptr) Get() interface{} {
	return vals.FromGo(reflect.Indirect(reflect.ValueOf(v.ptr)).Interface())
}

// Get sets the value pointed by the pointer, after conversion using ScanToGo.
func (v ptr) Set(val interface{}) error {
	return vals.ScanToGo(val, v.ptr)
}

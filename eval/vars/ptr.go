package vars

import (
	"reflect"
	"sync"

	"github.com/elves/elvish/eval/vals"
)

type ptrVariable struct {
	ptr   interface{}
	mutex *sync.RWMutex
}

// FromPtr creates a variable from a pointer. The variable is kept in sync
// with the value the pointer points to, using elvToGo and goToElv conversions
// when Get and Set. Its access is guarded by the supplied mutex.
func FromPtrWithMutex(p interface{}, m *sync.RWMutex) Var {
	return ptrVariable{p, m}
}

// FromPtr creates a variable from a pointer. The variable is kept in sync
// with the value the pointer points to, using elvToGo and goToElv conversions
// when Get and Set. Its access is guarded by a new mutex.
func FromPtr(p interface{}) Var {
	return FromPtrWithMutex(p, new(sync.RWMutex))
}

// FromInit creates a variable with an initial value. The variable created
// can be assigned values of any type.
func FromInit(v interface{}) Var {
	return FromPtr(&v)
}

// Get returns the value pointed by the pointer, after conversion using FromGo.
func (v ptrVariable) Get() interface{} {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return vals.FromGo(reflect.Indirect(reflect.ValueOf(v.ptr)).Interface())
}

// Get sets the value pointed by the pointer, after conversion using ScanToGo.
func (v ptrVariable) Set(val interface{}) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return vals.ScanToGo(val, v.ptr)
}

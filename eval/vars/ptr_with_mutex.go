package vars

import (
	"reflect"
	"sync"

	"github.com/elves/elvish/eval/vals"
)

type ptrWithMutex struct {
	ptr   interface{}
	mutex *sync.RWMutex
}

// FromPtr creates a variable from a pointer. The variable is kept in sync
// with the value the pointer points to, using elvToGo and goToElv conversions
// when Get and Set. Its access is guarded by the supplied mutex.
func FromPtrWithRWMutex(p interface{}, m *sync.RWMutex) Var {
	return ptr{p, m}
}

// Get returns the value pointed by the pointer, after conversion using FromGo.
func (v ptrWithMutex) Get() interface{} {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return vals.FromGo(reflect.Indirect(reflect.ValueOf(v.ptr)).Interface())
}

// Get sets the value pointed by the pointer, after conversion using ScanToGo.
func (v ptrWithMutex) Set(val interface{}) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return vals.ScanToGo(val, v.ptr)
}

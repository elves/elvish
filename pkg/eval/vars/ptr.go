package vars

import (
	"reflect"
	"sync"

	"src.elv.sh/pkg/eval/vals"
)

type PtrVar struct {
	ptr   any
	mutex *sync.RWMutex
}

// FromPtrWithMutex creates a variable from a pointer. The variable is kept in
// sync with the value the pointer points to, converting with vals.ScanToGo and
// vals.FromGo when Get and Set. Its access is guarded by the supplied mutex.
func FromPtrWithMutex(p any, m *sync.RWMutex) PtrVar {
	return PtrVar{p, m}
}

// FromPtr creates a variable from a pointer. The variable is kept in sync with
// the value the pointer points to, converting with vals.ScanToGo and
// vals.FromGo when Get and Set. Its access is guarded by a new mutex.
func FromPtr(p any) PtrVar {
	return FromPtrWithMutex(p, new(sync.RWMutex))
}

// FromInit creates a variable with an initial value. The variable created
// can be assigned values of any type.
func FromInit(v any) Var {
	return FromPtr(&v)
}

// Get returns the value pointed by the pointer, after conversion using FromGo.
func (v PtrVar) Get() any {
	return vals.FromGo(v.GetRaw())
}

// GetRaw returns the value pointed by the pointer without any conversion.
func (v PtrVar) GetRaw() any {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return reflect.Indirect(reflect.ValueOf(v.ptr)).Interface()
}

// Set sets the value pointed by the pointer, after conversion using ScanToGo.
func (v PtrVar) Set(val any) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return vals.ScanToGo(val, v.ptr)
}

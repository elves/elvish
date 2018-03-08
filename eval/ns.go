package eval

import (
	"fmt"
	"reflect"

	"github.com/elves/elvish/eval/vars"
)

// Ns is a map from names to variables.
type Ns map[string]vars.Var

var _ interface{} = Ns(nil)

// NewNs creates an empty namespace.
func NewNs() Ns {
	return make(Ns)
}

func (Ns) Kind() string {
	return "ns"
}

func (ns Ns) Hash() uint32 {
	return uint32(addrOf(ns))
}

func (ns Ns) Equal(rhs interface{}) bool {
	if ns2, ok := rhs.(Ns); ok {
		return addrOf(ns) == addrOf(ns2)
	}
	return false
}

func (ns Ns) Repr(int) string {
	return fmt.Sprintf("<ns 0x%x>", addrOf(ns))
}

func (ns Ns) Index(k interface{}) (interface{}, bool) {
	if kstring, ok := k.(string); ok {
		if v, ok := ns[kstring]; ok {
			return v.Get(), true
		}
	}
	return nil, false
}

func (ns Ns) IterateKeys(f func(interface{}) bool) {
	for k := range ns {
		if !f(k) {
			break
		}
	}
}

// HasName reports the namespace contains the given name.
func (ns Ns) HasName(name string) bool {
	_, ok := ns[name]
	return ok
}

// PopName removes a name from the namespace and returns the variable it used to
// contain.
func (ns Ns) PopName(name string) vars.Var {
	v := ns[name]
	delete(ns, name)
	return v
}

// Clone returns a shallow copy of the namespace.
func (ns Ns) Clone() Ns {
	ns2 := make(Ns)
	for name, variable := range ns {
		ns2[name] = variable
	}
	return ns2
}

// Add adds a variable to the namespace and returns the namespace itself.
func (ns Ns) Add(name string, v vars.Var) Ns {
	ns[name] = v
	return ns
}

// AddFn adds a function to a namespace. It returns the namespace itself.
func (ns Ns) AddFn(name string, v Callable) Ns {
	return ns.Add(name+FnSuffix, vars.FromPtr(&v))
}

// AddNs adds a sub-namespace to a namespace. It returns the namespace itself.
func (ns Ns) AddNs(name string, v Ns) Ns {
	return ns.Add(name+NsSuffix, vars.FromPtr(&v))
}

// AddBuiltinFn adds a builtin function to a namespace. It returns the namespace
// itself.
func (ns Ns) AddBuiltinFn(nsName, name string, impl interface{}) Ns {
	return ns.AddFn(name, NewBuiltinFn(nsName+name, impl))
}

// AddBuiltinFns adds builtin functions to a namespace. It returns the namespace
// itself.
func (ns Ns) AddBuiltinFns(nsName string, fns map[string]interface{}) Ns {
	for name, impl := range fns {
		ns.AddBuiltinFn(nsName, name, impl)
	}
	return ns
}

func addrOf(a interface{}) uintptr {
	return reflect.ValueOf(a).Pointer()
}

func (ns Ns) static() staticNs {
	static := make(staticNs)
	for name := range ns {
		static.set(name)
	}
	return static
}

// staticNs represents static information of an Ns.
type staticNs map[string]struct{}

func (ns staticNs) set(name string) {
	ns[name] = struct{}{}
}

func (ns staticNs) del(name string) {
	delete(ns, name)
}

func (ns staticNs) has(name string) bool {
	_, ok := ns[name]
	return ok
}

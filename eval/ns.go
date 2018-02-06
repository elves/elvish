package eval

import (
	"fmt"
	"reflect"

	"github.com/elves/elvish/eval/vartypes"
)

// Ns is a map from names to variables.
type Ns map[string]vartypes.Variable

var _ interface{} = Ns(nil)

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

func (ns Ns) Get(k interface{}) (interface{}, bool) {
	if kstring, ok := k.(string); ok {
		if v, ok := ns[kstring]; ok {
			return v.Get(), true
		}
	}
	return nil, false
}

func (ns Ns) SetFn(name string, v Callable) {
	ns[name+FnSuffix] = NewVariableFromPtr(&v)
}

func (ns Ns) SetNs(name string, v Ns) {
	ns[name+NsSuffix] = NewVariableFromPtr(&v)
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

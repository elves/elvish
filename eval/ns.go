package eval

import (
	"fmt"
	"reflect"

	"github.com/elves/elvish/eval/types"
)

// Ns is a map from names to variables.
type Ns map[string]Variable

var _ types.Value = Ns(nil)

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

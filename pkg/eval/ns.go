package eval

import (
	"fmt"
	"unsafe"

	"github.com/xiaq/persistent/hash"
	"src.elv.sh/pkg/eval/vars"
)

// Ns is the runtime representation of a namespace. The zero value of Ns is an
// empty namespace. To create a non-empty Ns, use either NsBuilder or CombineNs.
//
// An Ns is immutable after creation.
type Ns struct {
	// All variables in the namespace. Static variable accesses are compiled
	// into indexed accesses into this slice.
	slots []vars.Var
	// Names for the variables, used for introspection. Typical real programs
	// only contain a small number of names in each namespace, in which case a
	// linear search in a slice is usually faster than map access.
	names []string
	// Whether the variable has been deleted. Deleted variables can still be
	// kept in the Ns since there might be a reference to them in a closure.
	// Shadowed variables are also considered deleted.
	deleted []bool
}

// CombineNs returns an *Ns that contains all the bindings from both ns1 and
// ns2. Names in ns2 takes precedence over those in ns1.
func CombineNs(ns1 *Ns, ns2 *Ns) *Ns {
	ns := &Ns{
		append([]vars.Var(nil), ns2.slots...),
		append([]string(nil), ns2.names...),
		append([]bool(nil), ns2.deleted...)}
	hasName := map[string]bool{}
	for i, name := range ns.names {
		if !ns.deleted[i] {
			hasName[name] = true
		}
	}
	for i, name := range ns1.names {
		if !ns1.deleted[i] && !hasName[name] {
			ns.slots = append(ns.slots, ns1.slots[i])
			ns.names = append(ns.names, name)
			ns.deleted = append(ns.deleted, false)
		}
	}
	return ns
}

// Kind returns "ns".
func (ns *Ns) Kind() string {
	return "ns"
}

// Hash returns a hash of the address of ns.
func (ns *Ns) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(ns))
}

// Equal returns whether rhs has the same identity as ns.
func (ns *Ns) Equal(rhs interface{}) bool {
	if ns2, ok := rhs.(*Ns); ok {
		return ns == ns2
	}
	return false
}

// Repr returns an opaque representation of the Ns showing its address.
func (ns *Ns) Repr(int) string {
	return fmt.Sprintf("<ns %p>", ns)
}

// Index looks up a variable with the given name, and returns its value if it
// exists. This is only used for introspection from Elvish code; for
// introspection from Go code, use IndexName.
func (ns *Ns) Index(k interface{}) (interface{}, bool) {
	if ks, ok := k.(string); ok {
		variable := ns.IndexName(ks)
		if variable == nil {
			return nil, false
		}
		return variable.Get(), true
	}
	return nil, false
}

// Index looks up a variable with the given name, and returns its value if it
// exists, or nil if it does not. This is the type-safe version of Index and is
// useful for introspection from Go code.
func (ns *Ns) IndexName(k string) vars.Var {
	i := ns.lookup(k)
	if i != -1 {
		return ns.slots[i]
	}
	return nil
}

func (ns *Ns) lookup(k string) int {
	for i, name := range ns.names {
		if name == k && !ns.deleted[i] {
			return i
		}
	}
	return -1
}

// IterateKeys produces the names of all the variables in this Ns.
func (ns *Ns) IterateKeys(f func(interface{}) bool) {
	for i, name := range ns.names {
		if ns.slots[i] == nil || ns.deleted[i] {
			continue
		}
		if !f(name) {
			break
		}
	}
}

// IterateNames produces the names of all variables in the Ns. It is the
// type-safe version of IterateKeys and is useful for introspection from Go
// code. It doesn't support breaking early.
func (ns *Ns) IterateNames(f func(string)) {
	for i, name := range ns.names {
		if ns.slots[i] != nil && !ns.deleted[i] {
			f(name)
		}
	}
}

// HasName reports whether the Ns has a variable with the given name.
func (ns *Ns) HasName(k string) bool {
	for i, name := range ns.names {
		if name == k && !ns.deleted[i] {
			return ns.slots[i] != nil
		}
	}
	return false
}

func (ns *Ns) static() *staticNs {
	return &staticNs{ns.names, ns.deleted}
}

// NsBuilder is a helper type used for building an Ns.
type NsBuilder map[string]vars.Var

// Add adds a variable.
func (nb NsBuilder) Add(name string, v vars.Var) NsBuilder {
	nb[name] = v
	return nb
}

// AddFn adds a function. The resulting variable will be read-only.
func (nb NsBuilder) AddFn(name string, v Callable) NsBuilder {
	return nb.Add(name+FnSuffix, vars.NewReadOnly(v))
}

// AddNs adds a sub-namespace. The resulting variable will be read-only.
func (nb NsBuilder) AddNs(name string, v *Ns) NsBuilder {
	return nb.Add(name+NsSuffix, vars.NewReadOnly(v))
}

// AddGoFn adds a Go function. The resulting variable will be read-only.
func (nb NsBuilder) AddGoFn(nsName, name string, impl interface{}) NsBuilder {
	return nb.AddFn(name, NewGoFn(nsName+name, impl))
}

// AddGoFns adds Go functions. The resulting variables will be read-only.
func (nb NsBuilder) AddGoFns(nsName string, fns map[string]interface{}) NsBuilder {
	for name, impl := range fns {
		nb.AddGoFn(nsName, name, impl)
	}
	return nb
}

// Build builds an Ns.
func (nb NsBuilder) Ns() *Ns {
	n := len(nb)
	ns := &Ns{make([]vars.Var, n), make([]string, n), make([]bool, n)}
	i := 0
	for name, variable := range nb {
		ns.slots[i] = variable
		ns.names[i] = name
		i++
	}
	return ns
}

// The compile-time representation of a namespace. Called "static" namespace
// since it contains information that are known without executing the code.
// The data structure itself, however, is not static, and gets mutated as the
// compiler gains more information about the namespace. The zero value of
// staticNs is an empty namespace.
type staticNs struct {
	names   []string
	deleted []bool
}

func (ns *staticNs) clone() *staticNs {
	return &staticNs{
		append([]string{}, ns.names...), append([]bool{}, ns.deleted...)}
}

func (ns *staticNs) del(k string) {
	if i := ns.lookup(k); i != -1 {
		ns.deleted[i] = true
	}
}

// Adds a name, shadowing any existing one.
func (ns *staticNs) add(k string) int {
	ns.del(k)
	return ns.addInner(k)
}

// Adds a name, assuming that it either doesn't exist yet or has been deleted.
func (ns *staticNs) addInner(k string) int {
	ns.names = append(ns.names, k)
	ns.deleted = append(ns.deleted, false)
	return len(ns.names) - 1
}

func (ns *staticNs) lookup(k string) int {
	for i, name := range ns.names {
		if name == k && !ns.deleted[i] {
			return i
		}
	}
	return -1
}

type staticUpNs struct {
	names []string
	// For each name, whether the upvalue comes from the immediate outer scope,
	// i.e. the local scope a lambda is evaluated in.
	local []bool
	// Index of the upvalue variable, either into the local scope (if
	// the corresponding value in local is true) or the up scope (if the
	// corresponding value in local is false).
	index []int
}

func (up *staticUpNs) add(k string, local bool, index int) int {
	for i, name := range up.names {
		if name == k {
			return i
		}
	}
	up.names = append(up.names, k)
	up.local = append(up.local, local)
	up.index = append(up.index, index)
	return len(up.names) - 1
}

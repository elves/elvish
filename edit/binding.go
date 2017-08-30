package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

func getBinding(mode string, k ui.Key) eval.CallableValue {
	bindings := keyBindings[mode]
	if bindings == nil {
		return nil
	}
	v, ok := bindings[k]
	if ok {
		return v
	}
	return bindings[ui.Default]
}

// BindingTable adapts a binding table to eval.IndexSetter, so that it can be
// manipulated in elvish script.
type BindingTable struct {
	inner map[ui.Key]eval.CallableValue
}

// Kind returns "map".
func (BindingTable) Kind() string {
	return "map"
}

func (bt BindingTable) Equal(a interface{}) bool {
	rhs, ok := a.(BindingTable)
	if !ok {
		return false
	}
	if len(bt.inner) != len(rhs.inner) {
		return false
	}
	for k, v := range bt.inner {
		if !v.Equal(rhs.inner[k]) {
			return false
		}
	}
	return true
}

func (bt BindingTable) Hash() uint32 {
	h := hash.DJBInit
	for k, v := range bt.inner {
		// TODO(xiaq): Use a more efficient implementation to derive a hash from
		// ui.Key.
		h = hash.DJBCombine(h, hash.String(k.String()))
		h = hash.DJBCombine(h, v.Hash())
	}
	return h
}

// Repr returns the representation of the binding table as if it were an
// ordinary map.
func (bt BindingTable) Repr(indent int) string {
	var builder eval.MapReprBuilder
	builder.Indent = indent
	for k, v := range bt.inner {
		builder.WritePair(parse.Quote(k.String()), indent+2, v.Repr(indent+2))
	}
	return builder.String()
}

// IndexOne returns the value with the specified map key. The map key is first
// converted into an internal Key struct.
func (bt BindingTable) IndexOne(idx eval.Value) eval.Value {
	return bt.inner[ui.ToKey(idx)]
}

// IndexSet sets the value with the specified map key. The map key is first
// converted into an internal Key struct. The set value must be a callable one,
// otherwise an error is thrown.
func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := ui.ToKey(idx)
	f, ok := v.(eval.CallableValue)
	if !ok {
		throwf("want function, got %s", v.Kind())
	}
	bt.inner[key] = f
}

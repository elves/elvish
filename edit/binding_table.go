package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

func getBinding(bindingVar eval.Variable, k ui.Key) eval.Fn {
	binding := bindingVar.Get().(BindingTable)
	switch {
	case binding.HasKey(k):
		return binding.IndexOne(k).(eval.Fn)
	case binding.HasKey(ui.Default):
		return binding.IndexOne(ui.Default).(eval.Fn)
	default:
		return nil
	}
}

// BindingTable is a special Map that converts its key to ui.Key and ensures
// that its values satisfy eval.CallableValue.
type BindingTable struct {
	eval.Map
}

// Repr returns the representation of the binding table as if it were an
// ordinary map keyed by strings.
func (bt BindingTable) Repr(indent int) string {
	var builder eval.MapReprBuilder
	builder.Indent = indent
	bt.Map.IteratePair(func(k, v eval.Value) bool {
		builder.WritePair(parse.Quote(k.(ui.Key).String()), indent+2, v.Repr(indent+2))
		return true
	})
	return builder.String()
}

// IndexOne converts the index to ui.Key and uses the IndexOne of the inner Map.
func (bt BindingTable) IndexOne(idx eval.Value) eval.Value {
	return bt.Map.IndexOne(ui.ToKey(idx))
}

func (bt BindingTable) get(k ui.Key) eval.Fn {
	return bt.Map.IndexOne(k).(eval.Fn)
}

// Assoc converts the index to ui.Key, ensures that the value is CallableValue,
// uses the Assoc of the inner Map and converts the result to a BindingTable.
func (bt BindingTable) Assoc(k, v eval.Value) eval.Value {
	key := ui.ToKey(k)
	f, ok := v.(eval.Fn)
	if !ok {
		throwf("want function, got %s", v.Kind())
	}
	return BindingTable{bt.Map.Assoc(key, f).(eval.Map)}
}

package edit

import (
	"errors"
	"sort"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hashmap"
)

var errValueShouldBeFn = errors.New("value should be function")

func getBinding(bindingVar vartypes.Variable, k ui.Key) eval.Callable {
	return bindingVar.Get().(BindingTable).getOrDefault(k)
}

// BindingTable is a special Map that converts its key to ui.Key and ensures
// that its values satisfy eval.CallableValue.
type BindingTable struct {
	hashmap.Map
}

var emptyBindingTable = BindingTable{types.EmptyMap}

// Repr returns the representation of the binding table as if it were an
// ordinary map keyed by strings.
func (bt BindingTable) Repr(indent int) string {
	var builder types.MapReprBuilder
	builder.Indent = indent

	var keys ui.Keys
	for it := bt.Map.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		keys = append(keys, k.(ui.Key))
	}
	sort.Sort(keys)

	for _, k := range keys {
		v, _ := bt.Map.Get(k)
		builder.WritePair(parse.Quote(k.String()), indent+2, types.Repr(v, indent+2))
	}

	return builder.String()
}

// Index converts the index to ui.Key and uses the Index of the inner Map.
func (bt BindingTable) Index(index interface{}) (interface{}, error) {
	return types.Index(bt.Map, ui.ToKey(index))
}

func (bt BindingTable) HasKey(k interface{}) bool {
	_, ok := bt.Map.Get(k)
	return ok
}

func (bt BindingTable) get(k ui.Key) eval.Callable {
	v, ok := bt.Map.Get(k)
	if !ok {
		panic("get called when key not present")
	}
	return v.(eval.Callable)
}

func (bt BindingTable) getOrDefault(k ui.Key) eval.Callable {
	switch {
	case bt.HasKey(k):
		return bt.get(k)
	case bt.HasKey(ui.Default):
		return bt.get(ui.Default)
	}
	return nil
}

// Assoc converts the index to ui.Key, ensures that the value is CallableValue,
// uses the Assoc of the inner Map and converts the result to a BindingTable.
func (bt BindingTable) Assoc(k, v interface{}) (interface{}, error) {
	key := ui.ToKey(k)
	f, ok := v.(eval.Callable)
	if !ok {
		return nil, errValueShouldBeFn
	}
	map2 := bt.Map.Assoc(key, f)
	return BindingTable{map2}, nil
}

// Dissoc converts the key to ui.Key and calls the Dissoc method of the inner
// map.
func (bt BindingTable) Dissoc(k interface{}) interface{} {
	return BindingTable{bt.Map.Without(ui.ToKey(k))}
}

func makeBindingTable(raw hashmap.Map) BindingTable {
	converted := types.EmptyMap
	for it := raw.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		f, ok := v.(eval.Callable)
		if !ok {
			throw(errValueShouldBeFn)
		}
		converted = converted.Assoc(ui.ToKey(k), f)
	}

	return BindingTable{converted}
}

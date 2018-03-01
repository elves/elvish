package eddefs

import (
	"errors"
	"sort"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hashmap"
)

var errValueShouldBeFn = errors.New("value should be function")

// BindingMap is a special Map that converts its key to ui.Key and ensures
// that its values satisfy eval.CallableValue.
type BindingMap struct {
	hashmap.Map
}

var EmptyBindingMap = BindingMap{vals.EmptyMap}

// Repr returns the representation of the binding table as if it were an
// ordinary map keyed by strings.
func (bt BindingMap) Repr(indent int) string {
	var builder vals.MapReprBuilder
	builder.Indent = indent

	var keys ui.Keys
	for it := bt.Map.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		keys = append(keys, k.(ui.Key))
	}
	sort.Sort(keys)

	for _, k := range keys {
		v, _ := bt.Map.Index(k)
		builder.WritePair(parse.Quote(k.String()), indent+2, vals.Repr(v, indent+2))
	}

	return builder.String()
}

// Index converts the index to ui.Key and uses the Index of the inner Map.
func (bt BindingMap) Index(index interface{}) (interface{}, error) {
	return vals.Index(bt.Map, ui.ToKey(index))
}

func (bt BindingMap) HasKey(k interface{}) bool {
	_, ok := bt.Map.Index(k)
	return ok
}

func (bt BindingMap) GetKey(k ui.Key) eval.Callable {
	v, ok := bt.Map.Index(k)
	if !ok {
		panic("get called when key not present")
	}
	return v.(eval.Callable)
}

func (bt BindingMap) GetOrDefault(k ui.Key) eval.Callable {
	switch {
	case bt.HasKey(k):
		return bt.GetKey(k)
	case bt.HasKey(ui.Default):
		return bt.GetKey(ui.Default)
	}
	return nil
}

// Assoc converts the index to ui.Key, ensures that the value is CallableValue,
// uses the Assoc of the inner Map and converts the result to a BindingTable.
func (bt BindingMap) Assoc(k, v interface{}) (interface{}, error) {
	key := ui.ToKey(k)
	f, ok := v.(eval.Callable)
	if !ok {
		return nil, errValueShouldBeFn
	}
	map2 := bt.Map.Assoc(key, f)
	return BindingMap{map2}, nil
}

// Dissoc converts the key to ui.Key and calls the Dissoc method of the inner
// map.
func (bt BindingMap) Dissoc(k interface{}) interface{} {
	return BindingMap{bt.Map.Dissoc(ui.ToKey(k))}
}

func MakeBindingMap(raw hashmap.Map) (BindingMap, error) {
	converted := vals.EmptyMap
	for it := raw.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		f, ok := v.(eval.Callable)
		if !ok {
			return EmptyBindingMap, errValueShouldBeFn
		}
		converted = converted.Assoc(ui.ToKey(k), f)
	}

	return BindingMap{converted}, nil
}

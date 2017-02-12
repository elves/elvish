package edit

import (
	"github.com/elves/elvish/edit/uitypes"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// BindingTable adapts a binding table to eval.IndexSetter, so that it can be
// manipulated in elvish script.
type BindingTable struct {
	inner map[uitypes.Key]eval.FnValue
}

func (BindingTable) Kind() string {
	return "map"
}

func (bt BindingTable) Repr(indent int) string {
	var builder eval.MapReprBuilder
	builder.Indent = indent
	for k, v := range bt.inner {
		builder.WritePair(parse.Quote(k.String()), v.Repr(eval.IncIndent(indent, 1)))
	}
	return builder.String()
}

func (bt BindingTable) IndexOne(idx eval.Value) eval.Value {
	return bt.inner[uitypes.ToKey(idx)]
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := uitypes.ToKey(idx)
	f, ok := v.(eval.FnValue)
	if !ok {
		throwf("want function, got %s", v.Kind())
	}
	bt.inner[key] = f
}

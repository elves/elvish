package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// BindingTable adapts a binding table to eval.IndexSetter, so that it can be
// manipulated in elvish script.
type BindingTable struct {
	inner map[ui.Key]eval.CallableValue
}

func (BindingTable) Kind() string {
	return "map"
}

func (bt BindingTable) Repr(indent int) string {
	var builder eval.MapReprBuilder
	builder.Indent = indent
	for k, v := range bt.inner {
		builder.WritePair(parse.Quote(k.String()), indent+2, v.Repr(indent+2))
	}
	return builder.String()
}

func (bt BindingTable) IndexOne(idx eval.Value) eval.Value {
	return bt.inner[ui.ToKey(idx)]
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := ui.ToKey(idx)
	f, ok := v.(eval.CallableValue)
	if !ok {
		throwf("want function, got %s", v.Kind())
	}
	bt.inner[key] = f
}

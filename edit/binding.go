package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// Errors thrown to Evaler.
var (
	ErrTakeNoArg       = errors.New("editor builtins take no arguments")
	ErrEditorInactive  = errors.New("editor inactive")
	ErrKeyMustBeString = errors.New("key must be string")
)

// BindingTable adapts a binding table to eval.IndexSetter, so that it can be
// manipulated in elvish script.
type BindingTable struct {
	inner map[Key]eval.FnValue
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
	return bt.inner[ToKey(idx)]
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := ToKey(idx)
	f, ok := v.(eval.FnValue)
	if !ok {
		throwf("want function, got %s", v.Kind())
	}
	bt.inner[key] = f
}

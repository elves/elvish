package edit

import (
	"errors"
	"fmt"

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
	inner map[Key]BoundFunc
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
	key := ToKey(idx)
	switch f := bt.inner[key].(type) {
	case Builtin:
		return eval.String(f.name)
	case FnAsBoundFunc:
		return f.Fn
	}
	throw(errors.New("bug"))
	panic("unreachable")
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := ToKey(idx)

	var f BoundFunc
	switch v := v.(type) {
	case eval.String:
		builtin, ok := builtinMap[string(v)]
		if !ok {
			throw(fmt.Errorf("no builtin named %s", v.Repr(eval.NoPretty)))
		}
		f = builtin
	case eval.FnValue:
		f = FnAsBoundFunc{v}
	default:
		throw(fmt.Errorf("bad function type %s", v.Kind()))
	}

	bt.inner[key] = f
}

// BuiltinAsFnValue adapts a Builtin to satisfy eval.FnValue, so that it can be
// called from elvish script.
type BuiltinAsFnValue struct {
	b  Builtin
	ed *Editor
}

var _ eval.FnValue = &BuiltinAsFnValue{}

func (*BuiltinAsFnValue) Kind() string {
	return "fn"
}

func (eb *BuiltinAsFnValue) Repr(int) string {
	return "<editor builtin " + eb.b.name + ">"
}

func (eb *BuiltinAsFnValue) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	if len(args) > 0 {
		throw(ErrTakeNoArg)
	}
	if !eb.ed.active {
		throw(ErrEditorInactive)
	}
	eb.b.impl(eb.ed)
}

// BoundFunc is a function bound to a key. It is either a Builtin or an
// FnAsBoundFunc.
type BoundFunc interface {
	eval.Reprer
	Call(ed *Editor)
}

func (b Builtin) Repr(int) string {
	return b.name
}

func (b Builtin) Call(ed *Editor) {
	b.impl(ed)
}

// FnAsBoundFunc adapts eval.Fn to BoundFunc, so that functions in elvish
// script can be bound to keys.
type FnAsBoundFunc struct {
	Fn eval.FnValue
}

func (c FnAsBoundFunc) Repr(indent int) string {
	return c.Fn.Repr(indent)
}

func (c FnAsBoundFunc) Call(ed *Editor) {
	ed.CallFn(c.Fn)
}

package edit

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/errutil"
	"github.com/elves/elvish/eval"
)

// Exposing editor functionalities as an elvish module.

var (
	ErrTakeNoArg      = errors.New("editor builtins take no arguments")
	ErrEditorInactive = errors.New("editor inactive")
)

func makeModule(ed *Editor) eval.Namespace {
	ns := eval.Namespace{}
	// Populate builtins.
	for _, b := range builtins {
		ns[eval.FnPrefix+b.name] = eval.NewPtrVariable(&EditBuiltin{b, ed})
	}
	// Populate binding tables in the variable $binding.
	// TODO Make binding specific to the Editor.
	binding := map[eval.Value]eval.Value{
		eval.String("insert"):     BindingTable{keyBindings[modeInsert]},
		eval.String("command"):    BindingTable{keyBindings[modeCommand]},
		eval.String("completion"): BindingTable{keyBindings[modeCompletion]},
		eval.String("navigation"): BindingTable{keyBindings[modeNavigation]},
		eval.String("history"):    BindingTable{keyBindings[modeHistory]},
	}
	ns["binding"] = eval.NewPtrVariable(eval.NewMap(binding))

	return ns
}

// BindingTable adapts a binding table to eval.IndexSetter.
type BindingTable struct {
	inner map[Key]Caller
}

func (BindingTable) Kind() string {
	return "map"
}

func (bt BindingTable) Repr() string {
	var builder eval.MapReprBuilder
	for k, v := range bt.inner {
		builder.WritePair(k.String(), v.Repr())
	}
	return builder.String()
}

func (bt BindingTable) Index(idx eval.Value) eval.Value {
	key := keyIndex(idx)
	switch f := bt.inner[key].(type) {
	case Builtin:
		return eval.String(f.name)
	case EvalCaller:
		return f.Caller
	}
	throw(errors.New("bug"))
	panic("unreachable")
}

func (bt BindingTable) IndexSet(idx, v eval.Value) {
	key := keyIndex(idx)

	var f Caller
	switch v := v.(type) {
	case eval.String:
		builtin, ok := builtinMap[string(v)]
		if !ok {
			throw(fmt.Errorf("no builtin named %s", v.Repr()))
		}
		f = builtin
	case eval.Caller:
		f = EvalCaller{v}
	default:
		throw(fmt.Errorf("bad function type %s", v.Kind()))
	}

	bt.inner[key] = f
}

func keyIndex(idx eval.Value) Key {
	skey, ok := idx.(eval.String)
	if !ok {
		throw(errKeyMustBeString)
	}
	key, err := parseKey(string(skey))
	if err != nil {
		throw(err)
	}
	return key
}

// Builtin adapts a Builtin to satisfy eval.Value and eval.Caller.
type EditBuiltin struct {
	b  Builtin
	ed *Editor
}

func (*EditBuiltin) Kind() string {
	return "fn"
}

func (eb *EditBuiltin) Repr() string {
	return "<editor builtin " + eb.b.name + ">"
}

func (eb *EditBuiltin) Call(ec *eval.EvalCtx, args []eval.Value) {
	if len(args) > 0 {
		throw(ErrTakeNoArg)
	}
	if !eb.ed.active {
		throw(ErrEditorInactive)
	}
	eb.b.impl(eb.ed)
}

func throw(e error) {
	errutil.Throw(e)
}

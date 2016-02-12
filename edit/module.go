package edit

import (
	"errors"

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
	for name, impl := range builtins {
		ns[eval.FnPrefix+name] = eval.NewPtrVariable(&EditBuiltin{name, impl, ed})
	}
	return ns
}

// Builtin adapts a builtin to satisfy eval.Value and eval.Caller.
type EditBuiltin struct {
	name string
	impl builtin
	ed   *Editor
}

func (*EditBuiltin) Kind() string {
	return "fn"
}

func (eb *EditBuiltin) Repr() string {
	return "<editor builtin " + eb.name + ">"
}

func (eb *EditBuiltin) Call(ec *eval.EvalCtx, args []eval.Value) {
	if len(args) > 0 {
		throw(ErrTakeNoArg)
	}
	if !eb.ed.active {
		throw(ErrEditorInactive)
	}
	eb.impl(eb.ed)
}

func throw(e error) {
	errutil.Throw(e)
}

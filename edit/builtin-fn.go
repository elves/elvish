package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
)

// Type for line editor builtins. The implementations reside in files for each
// mode. For instance, builtins related to the location mode are in location.go.

var ErrEditorInactive = errors.New("editor inactive")

// BuiltinFn records an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.CallableValue = &BuiltinFn{}

func (*BuiltinFn) Kind() string {
	return "fn"
}

func (bf *BuiltinFn) Repr(int) string {
	return "$" + bf.name
}

func (bf *BuiltinFn) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok || !ed.active {
		throw(ErrEditorInactive)
	}
	bf.impl(ed)
}

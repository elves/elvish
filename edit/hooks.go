package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

// The $le:{before,after}-readline lists that contain hooks. We might have more
// hooks in future.

var _ = RegisterVariable("before-readline", makeListVariable)

func (ed *Editor) beforeReadLine() types.List {
	return ed.variables["before-readline"].Get().(types.List)
}

var _ = RegisterVariable("after-readline", makeListVariable)

func (ed *Editor) afterReadLine() types.List {
	return ed.variables["after-readline"].Get().(types.List)
}

func makeListVariable() vartypes.Variable {
	return vartypes.NewValidatedPtr(types.EmptyList, vartypes.ShouldBeList)
}

func callHooks(ev *eval.Evaler, li types.List, args ...types.Value) {
	if li.Len() == 0 {
		return
	}

	li.Iterate(func(v types.Value) bool {
		op := eval.Op{&hookOp{v, args}, -1, -1}
		ev.Eval(op, eval.NewInternalSource("[hooks]"))
		return true
	})
}

type hookOp struct {
	hook types.Value
	args []types.Value
}

func (op *hookOp) Invoke(fm *eval.Frame) error {
	fn, ok := op.hook.(eval.Fn)
	if !ok {
		fmt.Fprintf(os.Stderr, "not a function: %s\n", op.hook.Repr(types.NoPretty))
		return nil
	}
	err := fm.PCall(fn, op.args, eval.NoOpts)
	if err != nil {
		// TODO Print stack trace.
		fmt.Fprintf(os.Stderr, "function error: %s\n", err.Error())
	}
	return nil
}

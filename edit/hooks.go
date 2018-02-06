package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/vector"
)

// The $edit:{before,after}-readline lists that contain hooks. We might have more
// hooks in future.

type editorHooks struct {
	beforeReadline vector.Vector
	afterReadline  vector.Vector
}

func init() {
	atEditorInit(func(ed *Editor) {
		ed.beforeReadline = types.EmptyList
		ed.variables["before-readline"] = eval.NewVariableFromPtr(&ed.beforeReadline)
		ed.afterReadline = types.EmptyList
		ed.variables["after-readline"] = eval.NewVariableFromPtr(&ed.afterReadline)
	})
}

func makeListVariable() vartypes.Variable {
	l := types.EmptyList
	return eval.NewVariableFromPtr(&l)
}

func callHooks(ev *eval.Evaler, li vector.Vector, args ...interface{}) {
	if li.Len() == 0 {
		return
	}

	for it := li.Iterator(); it.HasElem(); it.Next() {
		op := eval.Op{&hookOp{it.Elem(), args}, -1, -1}
		ev.Eval(op, eval.NewInternalSource("[hooks]"))
	}
}

type hookOp struct {
	hook interface{}
	args []interface{}
}

func (op *hookOp) Invoke(fm *eval.Frame) error {
	fn, ok := op.hook.(eval.Callable)
	if !ok {
		fmt.Fprintf(os.Stderr, "not a function: %s\n", types.Repr(op.hook, types.NoPretty))
		return nil
	}
	err := fm.PCall(fn, op.args, eval.NoOpts)
	if err != nil {
		// TODO Print stack trace.
		fmt.Fprintf(os.Stderr, "function error: %s\n", err.Error())
	}
	return nil
}

package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/eval"
)

// The $le:{before,after}-readline lists that contain hooks. We might have more
// hooks in future.

var _ = RegisterVariable("before-readline", makeListVariable)

func (ed *Editor) beforeReadLine() eval.List {
	return ed.variables["before-readline"].Get().(eval.List)
}

var _ = RegisterVariable("after-readline", makeListVariable)

func (ed *Editor) afterReadLine() eval.List {
	return ed.variables["after-readline"].Get().(eval.List)
}

func makeListVariable() eval.Variable {
	return eval.NewPtrVariableWithValidator(eval.NewList(), eval.ShouldBeList)
}

func callHooks(ev *eval.Evaler, li eval.List, args ...eval.Value) {
	if li.Len() == 0 {
		return
	}

	li.Iterate(func(v eval.Value) bool {
		opfunc := func(ec *eval.Frame) {
			fn, ok := v.(eval.CallableValue)
			if !ok {
				fmt.Fprintf(os.Stderr, "not a function: %s\n", v.Repr(eval.NoPretty))
				return
			}
			err := ec.PCall(fn, args, eval.NoOpts)
			if err != nil {
				// TODO Print stack trace.
				fmt.Fprintf(os.Stderr, "function error: %s\n", err.Error())
			}
		}
		ev.Eval(eval.Op{opfunc, -1, -1}, "[hooks]", "no source")
		return true
	})
}

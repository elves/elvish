package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/eval"
)

var _ = registerVariable("before-readline", makeListVariable)

func (ed *Editor) beforeReadLine() eval.List {
	return ed.variables["before-readline"].Get().(eval.List)
}

var _ = registerVariable("after-readline", makeListVariable)

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

	opfunc := func(ec *eval.EvalCtx) {
		li.Iterate(func(v eval.Value) bool {
			fn, ok := v.(eval.CallableValue)
			if !ok {
				fmt.Fprintf(os.Stderr, "not a function: %s\n", v.Repr(eval.NoPretty))
				return true
			}
			err := ec.PCall(fn, args, eval.NoOpts)
			if err != nil {
				// TODO Print stack trace.
				fmt.Fprintf(os.Stderr, "function error: %s\n", err.Error())
			}
			return true
		})
	}
	ev.Eval(eval.Op{opfunc, -1, -1}, "[hooks]", "no source")
}

package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/vector"
)

// The $edit:{before,after}-readline lists that contain hooks. We might have more
// hooks in future.

// editorHooks contain hooks for the editor. They are just slices of functions;
// each of them is initialized with a function that calls all Elvish functions
// contained in the eponymous variable under edit:.
type editorHooks struct {
	beforeReadline []func()
	afterReadline  []func(string)
}

func init() {
	atEditorInit(func(ed *Editor, ns eval.Ns) {
		beforeReadline := types.EmptyList
		ns["before-readline"] = eval.NewVariableFromPtr(&beforeReadline)
		ed.beforeReadline = []func(){func() { callHooks(ed, beforeReadline) }}

		afterReadline := types.EmptyList
		ns["after-readline"] = eval.NewVariableFromPtr(&afterReadline)
		ed.afterReadline = []func(string){
			func(s string) { callHooks(ed, afterReadline, s) }}
	})
}

func callHooks(ed *Editor, li vector.Vector, args ...interface{}) {
	for it := li.Iterator(); it.HasElem(); it.Next() {
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO More detailed error message.
			ed.Notify("hook not a function")
		}
		ed.CallFn(fn, args...)
	}
}

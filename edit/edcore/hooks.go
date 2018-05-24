package edcore

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
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
	atEditorInit(func(ed *editor, ns eval.Ns) {
		beforeReadline := vals.EmptyList
		ns["before-readline"] = vars.FromPtr(&beforeReadline)
		ed.AddBeforeReadline(func() { callHooks(ed, beforeReadline) })

		afterReadline := vals.EmptyList
		ns["after-readline"] = vars.FromPtr(&afterReadline)
		ed.AddAfterReadline(func(s string) { callHooks(ed, afterReadline, s) })
	})
}

// AddBeforeReadline adds a function to the before-readline hook.
func (h *editorHooks) AddBeforeReadline(f func()) {
	h.beforeReadline = append(h.beforeReadline, f)
}

// AddAfterReadline adds a function to the after-readline hook.
func (h *editorHooks) AddAfterReadline(f func(string)) {
	h.afterReadline = append(h.afterReadline, f)
}

func callHooks(ed *editor, li vector.Vector, args ...interface{}) {
	for it := li.Iterator(); it.HasElem(); it.Next() {
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO More detailed error message.
			ed.Notify("hook not a function")
		}
		ed.CallFn(fn, args...)
	}
}

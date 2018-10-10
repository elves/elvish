package newedit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
)

func initBeforeReadline(ev *eval.Evaler) (vars.Var, func()) {
	hook := vals.EmptyList
	return vars.FromPtr(&hook), func() {
		i := -1
		for it := hook.Iterator(); it.HasElem(); it.Next() {
			i++
			name := fmt.Sprintf("$before-readline[%d]", i)
			fn, ok := it.Elem().(eval.Callable)
			if !ok {
				// TODO(xiaq): This is not testable as it depends on stderr.
				// Make it testable.
				diag.Complainf("%s not function", name)
				continue
			}
			// TODO(xiaq): This should use stdPorts, but stdPorts is currently
			// unexported from eval.
			ports := []*eval.Port{
				{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
			fm := eval.NewTopFrame(ev, eval.NewInternalSource(name), ports)
			fm.Call(fn, eval.NoArgs, eval.NoOpts)
		}
	}
}

func initAfterReadline(ev *eval.Evaler) (vars.Var, func(string)) {
	hook := vals.EmptyList
	return vars.FromPtr(&hook), func(code string) {
		i := -1
		for it := hook.Iterator(); it.HasElem(); it.Next() {
			i++
			name := fmt.Sprintf("$after-readline[%d]", i)
			fn, ok := it.Elem().(eval.Callable)
			if !ok {
				// TODO(xiaq): This is not testable as it depends on stderr.
				// Make it testable.
				diag.Complainf("%s not function", name)
				continue
			}
			// TODO(xiaq): This should use stdPorts, but stdPorts is currently
			// unexported from eval.
			ports := []*eval.Port{
				{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
			fm := eval.NewTopFrame(ev, eval.NewInternalSource(name), ports)
			fm.Call(fn, []interface{}{code}, eval.NoOpts)
		}
	}
}

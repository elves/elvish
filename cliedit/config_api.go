package cliedit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
)

func initConfigAPI(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	initMaxHeight(appSpec, ns)
	initBeforeReadline(appSpec, ev, ns)
	initAfterReadline(appSpec, ev, ns)
}

func initMaxHeight(appSpec *cli.AppSpec, ns eval.Ns) {
	maxHeight := newIntVar(-1)
	appSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	ns.Add("max-height", maxHeight)
}

func initBeforeReadline(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["before-readline"] = hook
	appSpec.BeforeReadline = append(appSpec.BeforeReadline, func() {
		callHooks(ev, "$<edit>:before-readline", hook.Get().(vals.List))
	})
}

func initAfterReadline(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["after-readline"] = hook
	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		callHooks(ev, "$<edit>:after-readline", hook.Get().(vals.List), code)
	})
}

func callHooks(ev *eval.Evaler, name string, hook vals.List, args ...interface{}) {
	i := -1
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
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
		fm.Call(fn, args, eval.NoOpts)
	}
}

func newIntVar(i int) vars.PtrVar            { return vars.FromPtr(&i) }
func newFloatVar(f float64) vars.PtrVar      { return vars.FromPtr(&f) }
func newBoolVar(b bool) vars.PtrVar          { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar     { return vars.FromPtr(&l) }
func newMapVar(m vals.Map) vars.PtrVar       { return vars.FromPtr(&m) }
func newFnVar(c eval.Callable) vars.PtrVar   { return vars.FromPtr(&c) }
func newBindingVar(b bindingMap) vars.PtrVar { return vars.FromPtr(&b) }

package edit

import (
	"fmt"
	"os"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/store/storedefs"
)

func initMaxHeight(appSpec *cli.AppSpec, nb eval.NsBuilder) {
	maxHeight := newIntVar(-1)
	appSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	nb.AddVar("max-height", maxHeight)
}

func initReadlineHooks(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	initBeforeReadline(appSpec, ev, nb)
	initAfterReadline(appSpec, ev, nb)
}

func initBeforeReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb.AddVar("before-readline", hook)
	appSpec.BeforeReadline = append(appSpec.BeforeReadline, func() {
		eval.CallHook(ev, nil, "$<edit>:before-readline", hook.Get().(vals.List))
	})
}

func initAfterReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb.AddVar("after-readline", hook)
	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		eval.CallHook(ev, nil, "$<edit>:after-readline", hook.Get().(vals.List), code)
	})
}

func initAddCmdFilters(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder, s histutil.Store) {
	ignoreLeadingSpace := eval.NewGoFn("<ignore-cmd-with-leading-space>",
		func(s string) bool { return !strings.HasPrefix(s, " ") })
	filters := newListVar(vals.MakeList(ignoreLeadingSpace))
	nb.AddVar("add-cmd-filters", filters)

	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		if code != "" &&
			callFilters(ev, "$<edit>:add-cmd-filters",
				filters.Get().(vals.List), code) {
			s.AddCmd(storedefs.Cmd{Text: code, Seq: -1})
		}
		// TODO(xiaq): Handle the error.
	})
}

func initGlobalBindings(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	appSpec.GlobalBindings = newMapBindings(nt, ev, bindingVar)
	nb.AddVar("global-binding", bindingVar)
}

func callFilters(ev *eval.Evaler, name string, filters vals.List, args ...any) bool {
	if filters.Len() == 0 {
		return true
	}

	i := -1
	for it := filters.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			complain("%s not function", name)
			continue
		}

		port1, collect, err := eval.ValueCapturePort()
		if err != nil {
			complain("cannot create pipe to run filter")
			return true
		}
		err = ev.Call(fn, eval.CallCfg{Args: args, From: name},
			// TODO: Supply the Chan component of port 2.
			eval.EvalCfg{Ports: []*eval.Port{nil, port1, {File: os.Stderr}}})
		out := collect()

		if err != nil {
			complain("%s return error", name)
			continue
		}
		if len(out) != 1 {
			complain("filter %s should only return $true or $false", name)
			continue
		}
		p, ok := out[0].(bool)
		if !ok {
			complain("filter %s should return bool", name)
			continue
		}
		if !p {
			return false
		}
	}
	return true
}

// TODO: This is not testable as it depends on stderr. Make it testable.
func complain(format string, args ...any) {
	diag.ShowError(os.Stderr, fmt.Errorf(format, args...))
}

func newIntVar(i int) vars.PtrVar             { return vars.FromPtr(&i) }
func newFloatVar(f float64) vars.PtrVar       { return vars.FromPtr(&f) }
func newBoolVar(b bool) vars.PtrVar           { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar      { return vars.FromPtr(&l) }
func newMapVar(m vals.Map) vars.PtrVar        { return vars.FromPtr(&m) }
func newFnVar(c eval.Callable) vars.PtrVar    { return vars.FromPtr(&c) }
func newBindingVar(b bindingsMap) vars.PtrVar { return vars.FromPtr(&b) }

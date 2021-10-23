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

//elvdoc:var max-height
//
// Maximum height the editor is allowed to use, defaults to `+Inf`.
//
// By default, the height of the editor is only restricted by the terminal
// height. Some modes like location mode can use a lot of lines; as a result,
// it can often occupy the entire terminal, and push up your scrollback buffer.
// Change this variable to a finite number to restrict the height of the editor.

func initMaxHeight(appSpec *cli.AppSpec, nb eval.NsBuilder) {
	maxHeight := newIntVar(-1)
	appSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	nb.AddVar("max-height", maxHeight)
}

func initReadlineHooks(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	initBeforeReadline(appSpec, ev, nb)
	initAfterReadline(appSpec, ev, nb)
}

//elvdoc:var before-readline
//
// A list of functions to call before each readline cycle. Each function is
// called without any arguments.

func initBeforeReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb.AddVar("before-readline", hook)
	appSpec.BeforeReadline = append(appSpec.BeforeReadline, func() {
		callHooks(ev, "$<edit>:before-readline", hook.Get().(vals.List))
	})
}

//elvdoc:var after-readline
//
// A list of functions to call after each readline cycle. Each function is
// called with a single string argument containing the code that has been read.

func initAfterReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb.AddVar("after-readline", hook)
	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		callHooks(ev, "$<edit>:after-readline", hook.Get().(vals.List), code)
	})
}

//elvdoc:var add-cmd-filters
//
// List of filters to run before adding a command to history.
//
// A filter is a function that takes a command as argument and outputs
// a boolean value. If any of the filters outputs `$false`, the
// command is not saved to history, and the rest of the filters are
// not run. The default value of this list contains a filter which
// ignores command starts with space.

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

//elvdoc:var global-binding
//
// Global keybindings, consulted for keys not handled by mode-specific bindings.
//
// See [Keybindings](#keybindings).

func initGlobalBindings(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	appSpec.GlobalBindings = newMapBindings(nt, ev, bindingVar)
	nb.AddVar("global-binding", bindingVar)
}

func callHooks(ev *eval.Evaler, name string, hook vals.List, args ...interface{}) {
	if hook.Len() == 0 {
		return
	}

	ports, cleanup := eval.PortsFromStdFiles(ev.ValuePrefix())
	evalCfg := eval.EvalCfg{Ports: ports[:]}
	defer cleanup()

	i := -1
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO(xiaq): This is not testable as it depends on stderr.
			// Make it testable.
			diag.Complainf(os.Stderr, "%s not function", name)
			continue
		}

		err := ev.Call(fn, eval.CallCfg{Args: args, From: name}, evalCfg)
		if err != nil {
			diag.ShowError(os.Stderr, err)
		}
	}
}

func callFilters(ev *eval.Evaler, name string, filters vals.List, args ...interface{}) bool {
	if filters.Len() == 0 {
		return true
	}

	i := -1
	for it := filters.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO(xiaq): This is not testable as it depends on stderr.
			// Make it testable.
			diag.Complainf(os.Stderr, "%s not function", name)
			continue
		}

		port1, collect, err := eval.CapturePort()
		if err != nil {
			diag.Complainf(os.Stderr, "cannot create pipe to run filter")
			return true
		}
		err = ev.Call(fn, eval.CallCfg{Args: args, From: name},
			// TODO: Supply the Chan component of port 2.
			eval.EvalCfg{Ports: []*eval.Port{nil, port1, {File: os.Stderr}}})
		out := collect()

		if err != nil {
			diag.Complainf(os.Stderr, "%s return error", name)
			continue
		}
		if len(out) != 1 {
			diag.Complainf(os.Stderr, "filter %s should only return $true or $false", name)
			continue
		}
		p, ok := out[0].(bool)
		if !ok {
			diag.Complainf(os.Stderr, "filter %s should return bool", name)
			continue
		}
		if !p {
			return false
		}
	}
	return true
}

func newIntVar(i int) vars.PtrVar             { return vars.FromPtr(&i) }
func newFloatVar(f float64) vars.PtrVar       { return vars.FromPtr(&f) }
func newBoolVar(b bool) vars.PtrVar           { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar      { return vars.FromPtr(&l) }
func newMapVar(m vals.Map) vars.PtrVar        { return vars.FromPtr(&m) }
func newFnVar(c eval.Callable) vars.PtrVar    { return vars.FromPtr(&c) }
func newBindingVar(b bindingsMap) vars.PtrVar { return vars.FromPtr(&b) }

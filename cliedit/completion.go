package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cliedit/complete"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

//elvdoc:var completion:binding
//
// Keybinding for the completion mode.

//elvdoc:fn completion:start
//
// Start the completion mode

func initCompletion(app *cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("completion",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:completion>", map[string]interface{}{
			"start": func() { completionStart(app, ev, binding) },
		}))
}

func completionStart(app *cli.App, ev *eval.Evaler, binding el.Handler) {
	buf := app.CodeArea.CopyState().CodeBuffer
	result, err := complete.Complete(
		complete.CodeBuffer{Content: buf.Content, Dot: buf.Dot},
		complete.Config{
			Filter:     complete.PrefixFilter,
			PureEvaler: pureEvaler{ev},
		})
	if err != nil {
		app.Notify(err.Error())
		return
	}
	completion.Start(app, completion.Config{
		Name: result.Name, Replace: result.Replace, Items: result.Items,
		Binding: binding})
}

type pureEvaler struct{ ev *eval.Evaler }

func (pureEvaler) EachExternal(f func(string)) { eval.EachExternal(f) }

func (pureEvaler) EachSpecial(f func(string)) {
	for name := range eval.IsBuiltinSpecial {
		f(name)
	}
}

func (pe pureEvaler) EachNs(f func(string)) { pe.ev.EachNsInTop(f) }

func (pe pureEvaler) EachVariableInNs(ns string, f func(string)) {
	pe.ev.EachVariableInTop(ns, f)
}

func (pe pureEvaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	return pe.ev.PurelyEvalPrimary(pn)
}

func (pe pureEvaler) PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return pe.ev.PurelyEvalCompound(cn)
}

func (pe pureEvaler) PurelyEvalPartialCompound(cn *parse.Compound, in *parse.Indexing) (string, error) {
	return pe.ev.PurelyEvalPartialCompound(cn, in)
}

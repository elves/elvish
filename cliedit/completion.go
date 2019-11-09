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

//elvdoc:fn complete-filename
//
// ```elvish
// edit:complete-filename @args
// ```
//
// Produces a list of filenames found in the directory of the last argument. All
// other arguments are ignored. If the last argument does not contain a path
// (either absolute or relative to the current directory), then the current
// directory is used. Relevant files are output as `edit:complex-candidate`
// objects.
//
// This function is the default handler for any commands without
// explicit handlers in `$edit:completion:arg-completer`. See [Argument
// Completer](#argument-completer).
//
// Example:
//
// ```elvish-transcript
// ~> edit:complete-filename ''
// ▶ (edit:complex-candidate Applications &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate Books &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate Desktop &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate Docsafe &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate Documents &code-suffix=/ &display-suffix='' &style='01;34')
// ...
// ~> edit:complete-filename .elvish/
// ▶ (edit:complex-candidate .elvish/aliases &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate .elvish/db &code-suffix=' ' &display-suffix='' &style='')
// ▶ (edit:complex-candidate .elvish/epm-installed &code-suffix=' ' &display-suffix='' &style='')
// ▶ (edit:complex-candidate .elvish/lib &code-suffix=/ &display-suffix='' &style='01;34')
// ▶ (edit:complex-candidate .elvish/rc.elv &code-suffix=' ' &display-suffix='' &style='')
// ```

//elvdoc:fn completion:start
//
// Start the completion mode.

//elvdoc:fn completion:close
//
// Closes the completion mode UI.

func initCompletion(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddGoFns("<edit>", map[string]interface{}{
		"complete-filename": wrapArgGenerator(complete.GenerateFileNames),
	})
	ns.AddNs("completion",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:completion>", map[string]interface{}{
			"start": func() { completionStart(app, ev, binding) },
			"close": func() { completion.Close(app) },
		}))
}

func completionStart(app cli.App, ev *eval.Evaler, binding el.Handler) {
	buf := app.CodeArea().CopyState().Buffer
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

type wrappedArgGenerator func(*eval.Frame, ...string) error

// Wraps an ArgGenerator into a function that can be then passed to
// eval.NewGoFn.
func wrapArgGenerator(gen complete.ArgGenerator) wrappedArgGenerator {
	return func(fm *eval.Frame, args ...string) error {
		rawCands, err := gen(args)
		if err != nil {
			return err
		}
		ch := fm.OutputChan()
		for _, rawCand := range rawCands {
			ch <- rawCand
		}
		return nil
	}
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

package cliedit

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cliedit/complete"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hash"
)

//elvdoc:var completion:binding
//
// Keybinding for the completion mode.

//elvdoc:fn complete-filename
//
// ```elvish
// edit:complete-filename $args...
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

//elvdoc:fn complex-candidate
//
// ```elvish
// edit:complex-candidate $stem &code-suffix='' &display-suffix=''
// ```

type complexCandidateOpts struct {
	CodeSuffix    string
	DisplaySuffix string
}

func (*complexCandidateOpts) SetDefaultOptions() {}

func complexCandidate(opts complexCandidateOpts, stem string) complexItem {
	return complexItem{
		Stem:          stem,
		CodeSuffix:    opts.CodeSuffix,
		DisplaySuffix: opts.DisplaySuffix,
	}
}

//elvdoc:fn match-prefix
//
// ```elvish
// edit:match-prefix $seed $inputs?
// ```
//
// For each input, outputs whether the input has $seed as a prefix. Uses the
// result of `to-string` for non-string inputs.
//
// Roughly equivalent to the following Elvish function, but more efficient:
//
// ```elvish
// use str
// fn match-prefix [seed @input]{
//   each [x]{ str:has-prefix (to-string $x) $seed } $@input
// }
// ```

//elvdoc:fn match-subseq
//
// ```elvish
// edit:match-subseq $seed $inputs?
// ```
//
// For each input, outputs whether the input has $seed as a
// [subsequence](https://en.wikipedia.org/wiki/Subsequence). Uses the result of
// `to-string` for non-string inputs.

//elvdoc:fn match-substr
//
// ```elvish
// edit:match-substr $seed $inputs?
// ```
//
// For each input, outputs whether the input has $seed as a substring. Uses the
// result of `to-string` for non-string inputs.
//
// Roughly equivalent to the following Elvish function, but more efficient:
//
// ```elvish
// use str
// fn match-substr [seed @input]{
//   each [x]{ str:has-contains (to-string $x) $seed } $@input
// }
// ```

//elvdoc:fn completion:start
//
// Start the completion mode.

func completionStart(app cli.App, ev *eval.Evaler, binding el.Handler) {
	buf := app.CodeArea().CopyState().Buffer
	result, err := complete.Complete(
		complete.CodeBuffer{Content: buf.Content, Dot: buf.Dot},
		complete.Config{
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

//elvdoc:fn completion:close
//
// Closes the completion mode UI.

func initCompletion(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddGoFns("<edit>", map[string]interface{}{
		"complete-filename": wrapArgGenerator(complete.GenerateFileNames),
		"complex-candidate": complexCandidate,
		"match-prefix":      wrapMatcher(strings.HasPrefix),
		"match-subseq":      wrapMatcher(util.HasSubseq),
		"match-substr":      wrapMatcher(strings.Contains),
	})
	ns.AddNs("completion",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:completion>", map[string]interface{}{
			"start": func() { completionStart(app, ev, binding) },
			"close": func() { completion.Close(app) },
		}))
}

// A wrapper type implementing Elvish value methods.
type complexItem complete.ComplexItem

func (c complexItem) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "stem":
		return c.Stem, true
	case "code-suffix":
		return c.CodeSuffix, true
	case "display-suffix":
		return c.DisplaySuffix, true
	}
	return nil, false
}

func (c complexItem) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, "stem", "code-suffix", "display-suffix")
}

func (c complexItem) Kind() string { return "map" }

func (c complexItem) Equal(a interface{}) bool {
	rhs, ok := a.(complexItem)
	return ok && c.Stem == rhs.Stem &&
		c.CodeSuffix == rhs.CodeSuffix && c.DisplaySuffix == rhs.DisplaySuffix
}

func (c complexItem) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.String(c.Stem))
	h = hash.DJBCombine(h, hash.String(c.CodeSuffix))
	h = hash.DJBCombine(h, hash.String(c.DisplaySuffix))
	return h
}

func (c complexItem) Repr(indent int) string {
	// TODO(xiaq): Pretty-print when indent >= 0
	return fmt.Sprintf("(edit:complex-candidate %s &code-suffix=%s &display-suffix=%s)",
		parse.Quote(c.Stem), parse.Quote(c.CodeSuffix), parse.Quote(c.DisplaySuffix))
}

type wrappedArgGenerator func(*eval.Frame, ...string) error

// Wraps an ArgGenerator into a function that can be then passed to
// eval.NewGoFn.
func wrapArgGenerator(gen complete.ArgGenerator) wrappedArgGenerator {
	return func(fm *eval.Frame, args ...string) error {
		rawItems, err := gen(args)
		if err != nil {
			return err
		}
		ch := fm.OutputChan()
		for _, rawItem := range rawItems {
			switch rawItem := rawItem.(type) {
			case complete.ComplexItem:
				ch <- complexItem(rawItem)
			case complete.PlainItem:
				ch <- string(rawItem)
			default:
				ch <- rawItem
			}
		}
		return nil
	}
}

// The type for a native Go matcher. This is not equivalent to the Elvish
// counterpart, which streams input and output. This is because we can actually
// afford calling a Go function for each item, so omitting the streaming
// behavior makes the implementation easier.
//
// Native Go matchers are wrapped into Elvish matchers, but never the other way
// around.
//
// This type is satisfied by strings.Contains and strings.HasPrefix; they are
// wrapped into match-substr and match-prefix respectively.
type matcher func(text, seed string) bool

type wrappedMatcher func(fm *eval.Frame, seed string, inputs eval.Inputs)

func wrapMatcher(m matcher) wrappedMatcher {
	return func(fm *eval.Frame, seed string, input eval.Inputs) {
		out := fm.OutputChan()
		input(func(v interface{}) {
			out <- m(vals.ToString(v), seed)
		})
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

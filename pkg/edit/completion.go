package edit

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/edit/complete"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/strutil"
)

//elvdoc:var completion:arg-completer
//
// A map containing argument completers.

//elvdoc:var completion:binding
//
// Keybinding for the completion mode.

//elvdoc:var completion:matcher
//
// A map mapping from context names to matcher functions. See the
// [Matcher](#matcher) section.

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
// ▶ (edit:complex-candidate Applications &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate Books &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate Desktop &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate Docsafe &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate Documents &code-suffix=/ &style='01;34')
// ...
// ~> edit:complete-filename .elvish/
// ▶ (edit:complex-candidate .elvish/aliases &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate .elvish/db &code-suffix=' ' &style='')
// ▶ (edit:complex-candidate .elvish/epm-installed &code-suffix=' ' &style='')
// ▶ (edit:complex-candidate .elvish/lib &code-suffix=/ &style='01;34')
// ▶ (edit:complex-candidate .elvish/rc.elv &code-suffix=' ' &style='')
// ```

//elvdoc:fn complex-candidate
//
// ```elvish
// edit:complex-candidate $stem &display='' &code-suffix=''
// ```
//
// Builds a complex candidate. This is mainly useful in [argument
// completers](#argument-completer).
//
// The `&display` option controls how the candidate is shown in the UI. It can
// be a string or a [styled](builtin.html#styled) text. If it is empty, `$stem`
// is used.
//
// The `&code-suffix` option affects how the candidate is inserted into the code
// when it is accepted. By default, a quoted version of `$stem` is inserted. If
// `$code-suffix` is non-empty, it is added to that text, and the suffix is not
// quoted.

type complexCandidateOpts struct {
	CodeSuffix string
	Display    string
}

func (*complexCandidateOpts) SetDefaultOptions() {}

func complexCandidate(fm *eval.Frame, opts complexCandidateOpts, stem string) complexItem {
	display := opts.Display
	if display == "" {
		display = stem
	}
	return complexItem{
		Stem:       stem,
		CodeSuffix: opts.CodeSuffix,
		Display:    display,
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

//elvdoc:fn completion:smart-start
//
// Starts the completion mode. However, if all the candidates share a non-empty
// prefix and that prefix starts with the seed, inserts the prefix instead.

func completionStart(app cli.App, bindings tk.Bindings, cfg complete.Config, smart bool) {
	buf := app.CodeArea().CopyState().Buffer
	result, err := complete.Complete(
		complete.CodeBuffer{Content: buf.Content, Dot: buf.Dot}, cfg)
	if err != nil {
		app.Notify(err.Error())
		return
	}
	if smart {
		prefix := ""
		for i, item := range result.Items {
			if i == 0 {
				prefix = item.ToInsert
				continue
			}
			prefix = commonPrefix(prefix, item.ToInsert)
			if prefix == "" {
				break
			}
		}
		if prefix != "" {
			insertedPrefix := false
			app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
				rep := s.Buffer.Content[result.Replace.From:result.Replace.To]
				if len(prefix) > len(rep) && strings.HasPrefix(prefix, rep) {
					s.Pending = tk.PendingCode{
						Content: prefix,
						From:    result.Replace.From, To: result.Replace.To}
					s.ApplyPending()
					insertedPrefix = true
				}
			})
			if insertedPrefix {
				return
			}
		}
	}
	w, err := mode.NewCompletion(app, mode.CompletionSpec{
		Name: result.Name, Replace: result.Replace, Items: result.Items,
		Filter: filterSpec, Bindings: bindings,
	})
	if w != nil {
		app.SetAddon(w, false)
	}
	if err != nil {
		app.Notify(err.Error())
	}
}

//elvdoc:fn completion:close
//
// Closes the completion mode UI.

func initCompletion(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	matcherMapVar := newMapVar(vals.EmptyMap)
	argGeneratorMapVar := newMapVar(vals.EmptyMap)
	cfg := func() complete.Config {
		return complete.Config{
			PureEvaler: pureEvaler{ev},
			Filterer: adaptMatcherMap(
				ed, ev, matcherMapVar.Get().(vals.Map)),
			ArgGenerator: adaptArgGeneratorMap(
				ev, argGeneratorMapVar.Get().(vals.Map)),
		}
	}
	generateForSudo := func(args []string) ([]complete.RawItem, error) {
		return complete.GenerateForSudo(cfg(), args)
	}
	nb.AddGoFns("<edit>", map[string]interface{}{
		"complete-filename": wrapArgGenerator(complete.GenerateFileNames),
		"complete-getopt":   completeGetopt,
		"complete-sudo":     wrapArgGenerator(generateForSudo),
		"complex-candidate": complexCandidate,
		"match-prefix":      wrapMatcher(strings.HasPrefix),
		"match-subseq":      wrapMatcher(strutil.HasSubseq),
		"match-substr":      wrapMatcher(strings.Contains),
	})
	app := ed.app
	nb.AddNs("completion",
		eval.NsBuilder{
			"arg-completer": argGeneratorMapVar,
			"binding":       bindingVar,
			"matcher":       matcherMapVar,
		}.AddGoFns("<edit:completion>:", map[string]interface{}{
			"accept":      func() { listingAccept(app) },
			"smart-start": func() { completionStart(app, bindings, cfg(), true) },
			"start":       func() { completionStart(app, bindings, cfg(), false) },
			"up":          func() { listingUp(app) },
			"down":        func() { listingDown(app) },
			"up-cycle":    func() { listingUpCycle(app) },
			"down-cycle":  func() { listingDownCycle(app) },
			"left":        func() { listingLeft(app) },
			"right":       func() { listingRight(app) },
		}).Ns())
}

// A wrapper type implementing Elvish value methods.
type complexItem complete.ComplexItem

func (c complexItem) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "stem":
		return c.Stem, true
	case "code-suffix":
		return c.CodeSuffix, true
	case "display":
		return c.Display, true
	}
	return nil, false
}

func (c complexItem) IterateKeys(f func(interface{}) bool) {
	vals.Feed(f, "stem", "code-suffix", "display")
}

// Kind is used by vals.Kind() to cause it to return the correct "kind" for these objects.
func (complexItem) Kind() string { return "map" }

func (c complexItem) Equal(a interface{}) bool {
	rhs, ok := a.(complexItem)
	return ok && c.Stem == rhs.Stem &&
		c.CodeSuffix == rhs.CodeSuffix && c.Display == rhs.Display
}

func (c complexItem) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.String(c.Stem))
	h = hash.DJBCombine(h, hash.String(c.CodeSuffix))
	h = hash.DJBCombine(h, hash.String(c.Display))
	return h
}

func (c complexItem) Repr(indent int) string {
	// TODO(xiaq): Pretty-print when indent >= 0
	return fmt.Sprintf("(edit:complex-candidate %s &code-suffix=%s &display=%s)",
		parse.Quote(c.Stem), parse.Quote(c.CodeSuffix), parse.Quote(c.Display))
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
		out := fm.ValueOutput()
		for _, rawItem := range rawItems {
			var v interface{}
			switch rawItem := rawItem.(type) {
			case complete.ComplexItem:
				v = complexItem(rawItem)
			case complete.PlainItem:
				v = string(rawItem)
			default:
				v = rawItem
			}
			err := out.Put(v)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func commonPrefix(s1, s2 string) string {
	for i, r := range s1 {
		if s2 == "" {
			break
		}
		r2, n2 := utf8.DecodeRuneInString(s2)
		if r2 != r {
			return s1[:i]
		}
		s2 = s2[n2:]
	}
	return s1
}

// The type for a native Go matcher. This is not equivalent to the Elvish
// counterpart, which streams input and output. This is because we can actually
// afford calling a Go function for each item, so omitting the streaming
// behavior makes the implementation simpler.
//
// Native Go matchers are wrapped into Elvish matchers, but never the other way
// around.
//
// This type is satisfied by strings.Contains and strings.HasPrefix; they are
// wrapped into match-substr and match-prefix respectively.
type matcher func(text, seed string) bool

type matcherOpts struct {
	IgnoreCase bool
	SmartCase  bool
}

func (*matcherOpts) SetDefaultOptions() {}

type wrappedMatcher func(fm *eval.Frame, opts matcherOpts, seed string, inputs eval.Inputs) error

func wrapMatcher(m matcher) wrappedMatcher {
	return func(fm *eval.Frame, opts matcherOpts, seed string, inputs eval.Inputs) error {
		out := fm.ValueOutput()
		var errOut error
		if opts.IgnoreCase || (opts.SmartCase && seed == strings.ToLower(seed)) {
			if opts.IgnoreCase {
				seed = strings.ToLower(seed)
			}
			inputs(func(v interface{}) {
				if errOut != nil {
					return
				}
				errOut = out.Put(m(strings.ToLower(vals.ToString(v)), seed))
			})
		} else {
			inputs(func(v interface{}) {
				if errOut != nil {
					return
				}
				errOut = out.Put(m(vals.ToString(v), seed))
			})
		}
		return errOut
	}
}

// Adapts $edit:completion:matcher into a Filterer.
func adaptMatcherMap(nt notifier, ev *eval.Evaler, m vals.Map) complete.Filterer {
	return func(ctxName, seed string, rawItems []complete.RawItem) []complete.RawItem {
		matcher, ok := lookupFn(m, ctxName)
		if !ok {
			nt.notifyf(
				"matcher for %s not a function, falling back to prefix matching", ctxName)
		}
		if matcher == nil {
			return complete.FilterPrefix(ctxName, seed, rawItems)
		}
		input := make(chan interface{})
		stopInputFeeder := make(chan struct{})
		defer close(stopInputFeeder)
		// Feed a string representing all raw candidates to the input channel.
		go func() {
			defer close(input)
			for _, rawItem := range rawItems {
				select {
				case input <- rawItem.String():
				case <-stopInputFeeder:
					return
				}
			}
		}()

		// TODO: Supply the Chan component of port 2.
		port1, collect, err := eval.CapturePort()
		if err != nil {
			nt.notifyf("cannot create pipe to run completion matcher: %v", err)
			return nil
		}

		err = ev.Call(matcher,
			eval.CallCfg{Args: []interface{}{seed}, From: "[editor matcher]"},
			eval.EvalCfg{Ports: []*eval.Port{
				// TODO: Supply the Chan component of port 2.
				{Chan: input, File: eval.DevNull}, port1, {File: os.Stderr}}})
		outputs := collect()

		if err != nil {
			nt.notifyError("matcher", err)
			// Continue with whatever values have been output
		}
		if len(outputs) != len(rawItems) {
			nt.notifyf(
				"matcher has output %v values, not equal to %v inputs",
				len(outputs), len(rawItems))
		}
		filtered := []complete.RawItem{}
		for i := 0; i < len(rawItems) && i < len(outputs); i++ {
			if vals.Bool(outputs[i]) {
				filtered = append(filtered, rawItems[i])
			}
		}
		return filtered
	}
}

func adaptArgGeneratorMap(ev *eval.Evaler, m vals.Map) complete.ArgGenerator {
	return func(args []string) ([]complete.RawItem, error) {
		gen, ok := lookupFn(m, args[0])
		if !ok {
			return nil, fmt.Errorf("arg completer for %s not a function", args[0])
		}
		if gen == nil {
			return complete.GenerateFileNames(args)
		}
		argValues := make([]interface{}, len(args))
		for i, arg := range args {
			argValues[i] = arg
		}
		var output []complete.RawItem
		var outputMutex sync.Mutex
		collect := func(item complete.RawItem) {
			outputMutex.Lock()
			defer outputMutex.Unlock()
			output = append(output, item)
		}
		valueCb := func(ch <-chan interface{}) {
			for v := range ch {
				switch v := v.(type) {
				case string:
					collect(complete.PlainItem(v))
				case complexItem:
					collect(complete.ComplexItem(v))
				default:
					collect(complete.PlainItem(vals.ToString(v)))
				}
			}
		}
		bytesCb := func(r *os.File) {
			buffered := bufio.NewReader(r)
			for {
				line, err := buffered.ReadString('\n')
				if line != "" {
					collect(complete.PlainItem(strutil.ChopLineEnding(line)))
				}
				if err != nil {
					break
				}
			}
		}
		port1, done, err := eval.PipePort(valueCb, bytesCb)
		if err != nil {
			panic(err)
		}
		err = ev.Call(gen,
			eval.CallCfg{Args: argValues, From: "[editor arg generator]"},
			eval.EvalCfg{Ports: []*eval.Port{
				// TODO: Supply the Chan component of port 2.
				nil, port1, {File: os.Stderr}}})
		done()

		return output, err
	}
}

func lookupFn(m vals.Map, ctxName string) (eval.Callable, bool) {
	val, ok := m.Index(ctxName)
	if !ok {
		val, ok = m.Index("")
	}
	if !ok {
		// No matcher, but not an error either
		return nil, true
	}
	fn, ok := val.(eval.Callable)
	if !ok {
		return nil, false
	}
	return fn, true
}

type pureEvaler struct{ ev *eval.Evaler }

func (pureEvaler) EachExternal(f func(string)) { fsutil.EachExternal(f) }

func (pureEvaler) EachSpecial(f func(string)) {
	for name := range eval.IsBuiltinSpecial {
		f(name)
	}
}

func (pe pureEvaler) EachNs(f func(string)) {
	eachNsInTop(pe.ev.Builtin(), pe.ev.Global(), f)
}

func (pe pureEvaler) EachVariableInNs(ns string, f func(string)) {
	eachVariableInTop(pe.ev.Builtin(), pe.ev.Global(), ns, f)
}

func (pe pureEvaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	return pe.ev.PurelyEvalPrimary(pn)
}

func (pe pureEvaler) PurelyEvalCompound(cn *parse.Compound) (string, bool) {
	return pe.ev.PurelyEvalCompound(cn)
}

func (pe pureEvaler) PurelyEvalPartialCompound(cn *parse.Compound, upto int) (string, bool) {
	return pe.ev.PurelyEvalPartialCompound(cn, upto)
}

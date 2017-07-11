package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
)

// The $edit:completer map, and its default values.

var (
	// ErrCompleterMustBeFn is thrown if the user has put a non-function entry
	// in $edit:completer, and that entry needs to be used for completion.
	// TODO(xiaq): Detect the type violation when the user modifies
	// $edit:completer.
	ErrCompleterMustBeFn = errors.New("completer must be fn")
	// ErrCompleterArgMustBeString is thrown when a builtin argument completer
	// is called with non-string arguments.
	ErrCompleterArgMustBeString = errors.New("arguments to arg completers must be string")
	// ErrTooFewArguments is thrown when a builtin argument completer is called
	// with too few arguments.
	ErrTooFewArguments = errors.New("too few arguments")
)

var (
	argCompletersData = map[string]*builtinArgCompleter{
		"":     {"complete-filename", complFilename},
		"sudo": {"complete-sudo", complSudo},
	}
)

var _ = registerVariable("arg-completer", argCompleterVariable)

func argCompleterVariable() eval.Variable {
	m := map[eval.Value]eval.Value{}
	for k, v := range argCompletersData {
		m[eval.String(k)] = v
	}
	return eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
}

func (ed *Editor) argCompleter() eval.Map {
	return ed.variables["arg-completer"].Get().(eval.Map)
}

func completeArg(words []string, ev *eval.Evaler) ([]rawCandidate, error) {
	logger.Printf("completing argument: %q", words)
	// XXX(xiaq): not the best way to get argCompleter.
	m := ev.Editor.(*Editor).argCompleter()
	var v eval.Value
	if m.HasKey(eval.String(words[0])) {
		v = m.IndexOne(eval.String(words[0]))
	} else {
		v = m.IndexOne(eval.String(""))
	}
	fn, ok := v.(eval.CallableValue)
	if !ok {
		return nil, ErrCompleterMustBeFn
	}
	return callArgCompleter(fn, ev, words)
}

type builtinArgCompleter struct {
	name string
	impl func([]string, *eval.Evaler) ([]rawCandidate, error)
}

var _ eval.CallableValue = &builtinArgCompleter{}

func (bac *builtinArgCompleter) Kind() string {
	return "fn"
}

func (bac *builtinArgCompleter) Repr(int) string {
	return "$edit:&" + bac.name
}

func (bac *builtinArgCompleter) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	words := make([]string, len(args))
	for i, arg := range args {
		s, ok := arg.(eval.String)
		if !ok {
			throw(ErrCompleterArgMustBeString)
		}
		words[i] = string(s)
	}
	cands, err := bac.impl(words, ec.Evaler)
	maybeThrow(err)
	out := ec.OutputChan()
	for _, cand := range cands {
		out <- cand
	}
}

func complFilename(words []string, ev *eval.Evaler) ([]rawCandidate, error) {
	if len(words) < 1 {
		return nil, ErrTooFewArguments
	}
	return complFilenameInner(words[len(words)-1], false)
}

func complSudo(words []string, ev *eval.Evaler) ([]rawCandidate, error) {
	if len(words) < 2 {
		return nil, ErrTooFewArguments
	}
	if len(words) == 2 {
		return complFormHeadInner(words[1], ev)
	}
	return completeArg(words[1:], ev)
}

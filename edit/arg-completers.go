package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
)

var (
	ErrCompleterMustBeFn        = errors.New("completer must be fn")
	ErrCompleterArgMustBeString = errors.New("arguments to arg completers must be string")
	ErrTooFewArguments          = errors.New("too few arguments")
)

var (
	argCompletersData = map[string]*builtinArgCompleter{
		"":     {"complete-filename", complFilename},
		"sudo": {"complete-sudo", complSudo},
	}
	argCompleter eval.Variable
)

func init() {
	m := map[eval.Value]eval.Value{}
	for k, v := range argCompletersData {
		m[eval.String(k)] = v
	}
	argCompleter = eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
}

func completeArg(words []string, ev *eval.Evaler) ([]*candidate, error) {
	logger.Printf("completing argument: %q", words)
	m := argCompleter.Get().(eval.Map)
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
	impl func([]string, *eval.Evaler) ([]*candidate, error)
}

var _ eval.CallableValue = &builtinArgCompleter{}

func (bac *builtinArgCompleter) Kind() string {
	return "fn"
}

func (bac *builtinArgCompleter) Repr(int) string {
	return "$le:&" + bac.name
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

func complFilename(words []string, ev *eval.Evaler) ([]*candidate, error) {
	if len(words) < 1 {
		return nil, ErrTooFewArguments
	}
	return complFilenameInner(words[len(words)-1], false)
}

func complSudo(words []string, ev *eval.Evaler) ([]*candidate, error) {
	if len(words) < 2 {
		return nil, ErrTooFewArguments
	}
	if len(words) == 2 {
		return complFormHeadInner(words[1], ev)
	}
	return completeArg(words[1:], ev)
}

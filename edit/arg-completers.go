package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
)

// CompleterTable provides $le:completer. It implements eval.IndexSetter.
type CompleterTable map[string]ArgCompleter

var _ eval.IndexSetter = CompleterTable(nil)

var (
	ErrCompleterIndexMustBeString = errors.New("index of completer table must be string")
	ErrCompleterValueMustBeFunc   = errors.New("value of completer table must be function")
)

func (CompleterTable) Kind() string {
	return "map"
}

func (ct CompleterTable) Repr(indent int) string {
	return "<repr not implemented yet>"
}

func (ct CompleterTable) IndexOne(idx eval.Value) eval.Value {
	head, ok := idx.(eval.String)
	if !ok {
		throw(ErrCompleterIndexMustBeString)
	}
	v := ct[string(head)]
	if fac, ok := v.(FnAsArgCompleter); ok {
		return fac.Fn
	}
	return eval.String("<get not implemented yet>")
}

func (ct CompleterTable) IndexSet(idx eval.Value, v eval.Value) {
	head, ok := idx.(eval.String)
	if !ok {
		throw(ErrCompleterIndexMustBeString)
	}
	value, ok := v.(eval.FnValue)
	if !ok {
		throw(ErrCompleterValueMustBeFunc)
	}
	ct[string(head)] = FnAsArgCompleter{value}
}

// ArgCompleter is an argument completer. Its Complete method is called with all
// words of the form. There are at least two words: the first one being the form
// head and the last word being the current argument to complete. It should
// return a list of candidates for the current argument and errors.
type ArgCompleter interface {
	Complete([]string, *eval.Evaler) ([]*candidate, error)
}

type FuncArgCompleter struct {
	impl func([]string, *eval.Evaler) ([]*candidate, error)
}

func (fac FuncArgCompleter) Complete(words []string, ev *eval.Evaler) ([]*candidate, error) {
	return fac.impl(words, ev)
}

var DefaultArgCompleter = ""
var argCompleter map[string]ArgCompleter

func init() {
	argCompleter = map[string]ArgCompleter{
		DefaultArgCompleter: FuncArgCompleter{complFilename},
		"sudo":              FuncArgCompleter{complSudo},
	}
}

func completeArg(words []string, ev *eval.Evaler) ([]*candidate, error) {
	Logger.Printf("completing argument: %q", words)
	compl, ok := argCompleter[words[0]]
	if !ok {
		compl = argCompleter[DefaultArgCompleter]
	}
	return compl.Complete(words, ev)
}

func complFilename(words []string, ev *eval.Evaler) ([]*candidate, error) {
	return complFilenameInner(words[len(words)-1], false)
}

func complFilenameFn(ec *eval.EvalCtx, word string) {
	cands, err := complFilenameInner(word, false)
	maybeThrow(err)
	out := ec.OutputChan()
	for _, cand := range cands {
		// TODO Preserve other parts of the candidate.
		out <- eval.String(cand.text)
	}
}

func complSudo(words []string, ev *eval.Evaler) ([]*candidate, error) {
	if len(words) == 2 {
		return complFormHeadInner(words[1], ev)
	}
	return completeArg(words[1:], ev)
}

type FnAsArgCompleter struct {
	Fn eval.FnValue
}

func (fac FnAsArgCompleter) Complete(words []string, ev *eval.Evaler) ([]*candidate, error) {
	return callArgCompleter(fac.Fn, ev, words)
}

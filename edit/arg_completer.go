package edit

import (
	"errors"
	"os"

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
	Complete([]string, *Editor) ([]*candidate, error)
}

type FuncArgCompleter struct {
	impl func([]string, *Editor) ([]*candidate, error)
}

func (fac FuncArgCompleter) Complete(words []string, ed *Editor) ([]*candidate, error) {
	return fac.impl(words, ed)
}

var DefaultArgCompleter = ""
var argCompleter map[string]ArgCompleter

func init() {
	argCompleter = map[string]ArgCompleter{
		DefaultArgCompleter: FuncArgCompleter{complFilename},
		"sudo":              FuncArgCompleter{complSudo},
	}
}

func completeArg(words []string, ed *Editor) ([]*candidate, error) {
	Logger.Printf("completing argument: %q", words)
	compl, ok := argCompleter[words[0]]
	if !ok {
		compl = argCompleter[DefaultArgCompleter]
	}
	return compl.Complete(words, ed)
}

func complFilename(words []string, ed *Editor) ([]*candidate, error) {
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

func complSudo(words []string, ed *Editor) ([]*candidate, error) {
	if len(words) == 2 {
		return complFormHeadInner(words[1], ed)
	}
	return completeArg(words[1:], ed)
}

type FnAsArgCompleter struct {
	Fn eval.FnValue
}

func (fac FnAsArgCompleter) Complete(words []string, ed *Editor) ([]*candidate, error) {
	in, err := makeClosedStdin()
	if err != nil {
		return nil, err
	}
	ports := []*eval.Port{in, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	wordValues := make([]eval.Value, len(words))
	for i, word := range words {
		wordValues[i] = eval.String(word)
	}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor completer]", "", ports)
	values, err := ec.PCaptureOutput(fac.Fn, wordValues)
	if err != nil {
		ed.notify("completer error: %v", err)
		return nil, err
	}

	cands := make([]*candidate, len(values))
	for i, v := range values {
		s := eval.ToString(v)
		cands[i] = &candidate{text: s}
	}
	return cands, nil
}

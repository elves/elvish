package edit

import (
	"errors"
	"os"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

var (
	errIncorrectNumOfResults = errors.New("matcher must return a bool for each candidate")
	errMatcherMustBeFn       = errors.New("matcher must be a function")
)

var prefixMatcher = eval.BuiltinFn{"prefixMatcher", func(ec *eval.EvalCtx,
	args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputChan()
	var pattern eval.String
	var candidates eval.List
	eval.ScanArgs(args, &pattern, &candidates)

	candidates.Iterate(func(cand eval.Value) bool {
		candidate := eval.ToString(cand)

		out <- eval.Bool(strings.HasPrefix(candidate, string(pattern)))
		return true
	})
}}

var _ = registerVariable("-matcher", func() eval.Variable {
	m := map[eval.Value]eval.Value{
		eval.String("index"):    eval.CallableValue(&prefixMatcher),
		eval.String("redirect"): eval.CallableValue(&prefixMatcher),
		eval.String("argument"): eval.CallableValue(&prefixMatcher),
		eval.String("variable"): eval.CallableValue(&prefixMatcher),
		eval.String("command"):  eval.CallableValue(&prefixMatcher),
	}
	return eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
})

func (ed *Editor) filterAndCookCandidates(ev *eval.Evaler, completer string, pattern string,
	cands []rawCandidate, q parse.PrimaryType) ([]*candidate, error) {

	var filtered []*candidate
	matcher := ed.variables["-matcher"].Get().(eval.Map).IndexOne(eval.String(completer))
	m, ok := matcher.(eval.CallableValue)
	if !ok {
		return nil, errMatcherMustBeFn
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan, {File: os.Stdout}, {File: os.Stderr}}
	ec := eval.NewTopEvalCtx(ev, "[editor matcher]", "", ports)

	var candValues []eval.Value
	patternValue := eval.String(pattern)

	for _, cand := range cands {
		candValues = append(candValues, eval.String(cand.text()))
	}

	args := []eval.Value{
		patternValue,
		eval.NewList(candValues...),
	}
	values, err := ec.PCaptureOutput(m, args, eval.NoOpts)
	if err != nil {
		return nil, err
	} else if len(values) != len(candValues) {
		return nil, errIncorrectNumOfResults
	}
	for i, value := range values {
		if eval.ToBool(value) {
			filtered = append(filtered, cands[i].cook(q))
		}
	}
	return filtered, nil
}

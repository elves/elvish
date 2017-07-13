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
	in := ec.InputChan()
	var pattern eval.String
	eval.ScanArgs(args, &pattern)

	for cand := range in {
		candidate := eval.ToString(cand)

		out <- eval.Bool(strings.HasPrefix(candidate, string(pattern)))
	}
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

	input := make(chan eval.Value, len(cands))
	ports := []*eval.Port{
		{Chan: input}, {File: os.Stdout}, {File: os.Stderr}}
	ec := eval.NewTopEvalCtx(ev, "[editor matcher]", "", ports)

	args := []eval.Value{
		eval.String(pattern),
	}

	for _, cand := range cands {
		input <- eval.String(cand.text())
	}
	close(input)

	values, err := ec.PCaptureOutput(m, args, eval.NoOpts)
	if err != nil {
		return nil, err
	} else if len(values) != len(cands) {
		return nil, errIncorrectNumOfResults
	}
	for i, value := range values {
		if eval.ToBool(value) {
			filtered = append(filtered, cands[i].cook(q))
		}
	}
	return filtered, nil
}

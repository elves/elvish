package edit

import (
	"errors"
	"strings"

	"github.com/elves/elvish/eval"
)

var (
	errIncorrectNumOfResults = errors.New("matcher must return a bool for each candidate")
	errMatcherMustBeFn       = errors.New("matcher must be a function")
)

func matchPrefix(ec *eval.EvalCtx,
	args []eval.Value, opts map[string]eval.Value) {

	var pattern eval.String
	iterate := eval.ScanArgsAndOptionalIterate(ec, args, &pattern)
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	iterate(func(cand eval.Value) {
		candidate := eval.ToString(cand)
		out <- eval.Bool(strings.HasPrefix(candidate, string(pattern)))
	})
}

var (
	matchPrefixFn = &eval.BuiltinFn{"match-prefix", matchPrefix}
	matchers      = []*eval.BuiltinFn{matchPrefixFn}

	_ = registerVariable("-matcher", func() eval.Variable {
		m := map[eval.Value]eval.Value{
			eval.String("index"):    matchPrefixFn,
			eval.String("redirect"): matchPrefixFn,
			eval.String("argument"): matchPrefixFn,
			eval.String("variable"): matchPrefixFn,
			eval.String("command"):  matchPrefixFn,
		}
		return eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
	})
)

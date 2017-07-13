package edit

import (
	"errors"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var (
	errIncorrectNumOfResults    = errors.New("matcher must return a bool for each candidate")
	errMatcherMustBeFn          = errors.New("matcher must be a function")
	errMatcherInputMustBeString = errors.New("matcher input must be string")
)

var (
	matchPrefix = &eval.BuiltinFn{
		"match-prefix", wrapMatcher(strings.HasPrefix)}
	matchSubseq = &eval.BuiltinFn{
		"match-subseq", wrapMatcher(util.HasSubseq)}
	matchers = []*eval.BuiltinFn{
		matchPrefix,
		matchSubseq,
	}

	_ = registerVariable("-matcher", func() eval.Variable {
		m := map[eval.Value]eval.Value{
			eval.String("index"):    matchPrefix,
			eval.String("redirect"): matchPrefix,
			eval.String("argument"): matchPrefix,
			eval.String("variable"): matchPrefix,
			eval.String("command"):  matchPrefix,
		}
		return eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
	})
)

func wrapMatcher(match func(s, p string) bool) eval.BuiltinFnImpl {
	return func(ec *eval.EvalCtx,
		args []eval.Value, opts map[string]eval.Value) {

		var pattern eval.String
		iterate := eval.ScanArgsAndOptionalIterate(ec, args, &pattern)
		eval.TakeNoOpt(opts)

		out := ec.OutputChan()
		iterate(func(v eval.Value) {
			s, ok := v.(eval.String)
			if !ok {
				throw(errMatcherInputMustBeString)
			}
			out <- eval.Bool(match(string(s), string(pattern)))
		})
	}
}

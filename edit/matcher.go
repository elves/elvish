package edit

import (
	"errors"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
)

var (
	errIncorrectNumOfResults    = errors.New("matcher must return a bool for each candidate")
	errMatcherMustBeFn          = errors.New("matcher must be a function")
	errMatcherInputMustBeString = errors.New("matcher input must be string")
)

var (
	matchPrefix = &eval.BuiltinFn{
		"edit:match-prefix", wrapMatcher(strings.HasPrefix)}
	matchSubstr = &eval.BuiltinFn{
		"edit:match-substr", wrapMatcher(strings.Contains)}
	matchSubseq = &eval.BuiltinFn{
		"edit:match-subseq", wrapMatcher(util.HasSubseq)}
	matchers = []*eval.BuiltinFn{
		matchPrefix,
		matchSubstring,
		matchSubseq,
	}

	_ = RegisterVariable("-matcher", func() eval.Variable {
		m := hashmap.Empty.Assoc(
			// Fallback matcher uses empty string as key
			eval.String(""), matchPrefix)
		return eval.NewPtrVariableWithValidator(eval.NewMap(m), eval.ShouldBeMap)
	})
)

func (ed *Editor) lookupMatcher(name string) (eval.CallableValue, bool) {
	m := ed.variables["-matcher"].Get().(eval.Map)
	if !m.HasKey(eval.String(name)) {
		// Use fallback matcher
		name = ""
	}
	matcher, ok := m.IndexOne(eval.String(name)).(eval.CallableValue)
	return matcher, ok
}

func wrapMatcher(matcher func(s, p string) bool) eval.BuiltinFnImpl {
	return func(ec *eval.EvalCtx,
		args []eval.Value, opts map[string]eval.Value) {

		var pattern eval.String
		iterate := eval.ScanArgsOptionalInput(ec, args, &pattern)
		var options struct {
			IgnoreCase bool
			SmartCase  bool
		}
		eval.ScanOptsToStruct(opts, &options)
		switch {
		case options.IgnoreCase && options.SmartCase:
			throwf("-ignore-case and -smart-case cannot be used together")
		case options.IgnoreCase:
			innerMatcher := matcher
			matcher = func(s, p string) bool {
				return innerMatcher(strings.ToLower(s), strings.ToLower(p))
			}
		case options.SmartCase:
			innerMatcher := matcher
			matcher = func(s, p string) bool {
				if p == strings.ToLower(p) {
					// Ignore case is pattern is all lower case.
					return innerMatcher(strings.ToLower(s), p)
				} else {
					return innerMatcher(s, p)
				}
			}
		}

		out := ec.OutputChan()
		iterate(func(v eval.Value) {
			s, ok := v.(eval.String)
			if !ok {
				throw(errMatcherInputMustBeString)
			}
			out <- eval.Bool(matcher(string(s), string(pattern)))
		})
	}
}

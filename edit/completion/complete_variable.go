package completion

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type variableComplContext struct {
	complContextCommon
	ns string
}

func (*variableComplContext) name() string { return "variable" }

func findVariableComplContext(n parse.Node, _ pureEvaler) complContext {
	primary, ok := n.(*parse.Primary)
	if ok && primary.Type == parse.Variable {
		sigil, qname := eval.SplitVariableRef(primary.Value)
		ns, nameSeed := eval.SplitQNameNsIncomplete(qname)

		// Move past "$", "@" and "<ns>:".
		begin := primary.Range().From + 1 + len(sigil) + len(ns)
		return &variableComplContext{
			complContextCommon{nameSeed, parse.Bareword, begin, primary.Range().To},
			ns,
		}
	}
	return nil
}

type evalerScopes interface {
	EachVariableInTop(string, func(string))
	EachNsInTop(func(string))
}

func (ctx *variableComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	complVariable(ctx.ns, env.evaler, ch)
	return nil
}

func complVariable(ctxNs string, ev evalerScopes, ch chan<- rawCandidate) {
	ev.EachVariableInTop(ctxNs, func(varname string) {
		ch <- noQuoteCandidate(varname)
	})

	ev.EachNsInTop(func(ns string) {
		// This is to match namespaces that are "nested" under the current
		// namespace.
		if hasProperPrefix(ns, ctxNs) {
			ch <- noQuoteCandidate(ns[len(ctxNs):])
		}
	})
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}

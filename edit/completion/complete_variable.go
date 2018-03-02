package completion

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type variableComplContext struct {
	complContextCommon
	ns, nsPart string
}

func (*variableComplContext) name() string { return "variable" }

func findVariableComplContext(n parse.Node, _ pureEvaler) complContext {
	primary := parse.GetPrimary(n)
	if primary != nil && primary.Type == parse.Variable {
		explode, nsPart, nameSeed := eval.SplitIncompleteVariableRef(primary.Value)
		// Move past "$", "@" and "<ns>:".
		begin := primary.Begin() + 1 + len(explode) + len(nsPart)
		ns := nsPart
		if len(ns) > 0 {
			ns = ns[:len(ns)-1]
		}
		return &variableComplContext{
			complContextCommon{nameSeed, parse.Bareword, begin, primary.End()},
			ns, nsPart,
		}
	}
	return nil
}

type evalerScopes interface {
	EachVariableInTop(string, func(string))
	EachNsInTop(func(string))
}

func (ctx *variableComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	complVariable(ctx.ns, ctx.nsPart, env.evaler, ch)
	return nil
}

func complVariable(ctxNs, ctxNsPart string, ev evalerScopes, ch chan<- rawCandidate) {
	ev.EachVariableInTop(ctxNs, func(varname string) {
		ch <- noQuoteCandidate(varname)
	})

	ev.EachNsInTop(func(ns string) {
		nsPart := ns + ":"
		// This is to match namespaces that are "nested" under the current
		// namespace.
		if hasProperPrefix(nsPart, ctxNsPart) {
			ch <- noQuoteCandidate(nsPart[len(ctxNsPart):])
		}
	})
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}

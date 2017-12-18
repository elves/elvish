package edit

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
		explode, qname := eval.ParseVariableSplice(primary.Value)
		nsPart, nameSeed := eval.ParseVariableQName(qname)
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

func (ctx *variableComplContext) generate(ev *eval.Evaler, ch chan<- rawCandidate) error {
	// Collect matching variables.
	ev.EachVariableInTop(ctx.ns, func(varname string) {
		ch <- noQuoteCandidate(varname)
	})

	ev.EachNsInTop(func(ns string) {
		nsPart := ns + ":"
		// This is to match namespaces that are "nested" under the current
		// namespace.
		if hasProperPrefix(nsPart, ctx.nsPart) {
			ch <- noQuoteCandidate(nsPart[len(ctx.nsPart):])
		}
	})

	return nil
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}

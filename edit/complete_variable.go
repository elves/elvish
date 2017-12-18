package edit

import (
	"os"
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
	iterateVariables(ev, ctx.ns, func(varname string) {
		ch <- noQuoteCandidate(varname)
	})

	seenMod := func(mod string) {
		modNsPart := mod + ":"
		// This is to match namespaces that are "nested" under the current
		// namespace.
		if hasProperPrefix(modNsPart, ctx.nsPart) {
			ch <- noQuoteCandidate(modNsPart[len(ctx.nsPart):])
		}
	}

	// Collect namespace prefixes.
	// TODO Support non-module namespaces.
	for mod := range ev.Global.Uses {
		seenMod(mod)
	}
	for mod := range ev.Builtin.Uses {
		seenMod(mod)
	}
	return nil
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}

// TODO: Make this a method of Evaler
func iterateVariables(ev *eval.Evaler, ns string, f func(string)) {
	switch ns {
	case "":
		for varname := range ev.Builtin.Names {
			f(varname)
		}
		for varname := range ev.Global.Names {
			f(varname)
		}
		// TODO Include local names as well.
	case "E":
		for _, s := range os.Environ() {
			f(s[:strings.IndexByte(s, '=')])
		}
	default:
		// TODO Support non-module namespaces.
		mod := ev.Global.Uses[ns]
		if mod == nil {
			mod = ev.Builtin.Uses[ns]
		}
		for varname := range mod {
			f(varname)
		}
	}
}

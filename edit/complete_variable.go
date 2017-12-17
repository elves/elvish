package edit

import (
	"os"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type variableCompleter struct {
	ns, nsPart string
	nameSeed   string
	begin, end int
}

func (*variableCompleter) name() string { return "variable" }

func findVariableCompleter(n parse.Node) completerIface {
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
		return &variableCompleter{ns, nsPart, nameSeed, begin, primary.End()}
	}
	return nil
}

func (compl *variableCompleter) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	rawCands := make(chan rawCandidate)
	go func() {
		defer close(rawCands)

		// Collect matching variables.
		iterateVariables(ev, compl.ns, func(varname string) {
			rawCands <- noQuoteCandidate(varname)
		})

		seenMod := func(mod string) {
			modNsPart := mod + ":"
			// This is to match namespaces that are "nested" under the current
			// namespace.
			if hasProperPrefix(modNsPart, compl.nsPart) {
				rawCands <- noQuoteCandidate(modNsPart[len(compl.nsPart):])
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
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(ev, matcher, compl.nameSeed,
		rawCands, parse.Bareword)
	if err != nil {
		return nil, err
	}

	return &complSpec{compl.begin, compl.end, cands}, nil
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

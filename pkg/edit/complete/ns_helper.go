package complete

import (
	"os"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
	"src.elv.sh/pkg/parse/np"
)

var environ = os.Environ

// Calls f for each variable name in namespace ns that can be found at the point
// of np.
func eachVariableInNs(ev *eval.Evaler, p np.Path, ns string, f func(s string)) {
	switch ns {
	case "", ":":
		ev.Global().IterateKeysString(f)
		ev.Builtin().IterateKeysString(f)
		eachDefinedVariable(p[len(p)-1], p[0].Range().From, f)
	case "e:":
		eachExternal(func(cmd string) {
			f(cmd + eval.FnSuffix)
		})
	case "E:":
		for _, s := range environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		// TODO: Support namespaces defined in the code too.
		segs := eval.SplitQNameSegs(ns)
		mod := ev.Global().IndexString(segs[0])
		if mod == nil {
			mod = ev.Builtin().IndexString(segs[0])
		}
		for _, seg := range segs[1:] {
			if mod == nil {
				return
			}
			mod = mod.Get().(*eval.Ns).IndexString(seg)
		}
		if mod != nil {
			mod.Get().(*eval.Ns).IterateKeysString(f)
		}
	}
}

// Calls f for each variables defined in n that are visible at pos.
func eachDefinedVariable(n parse.Node, pos int, f func(string)) {
	if fn, ok := n.(*parse.Form); ok {
		eachDefinedVariableInForm(fn, f)
	}
	if pn, ok := n.(*parse.Primary); ok && pn.Type == parse.Lambda {
		for _, param := range pn.Elements {
			if varRef, ok := cmpd.StringLiteral(param); ok {
				_, name := eval.SplitSigil(varRef)
				f(name)
			}
		}
	}
	for _, ch := range parse.Children(n) {
		if ch.Range().From > pos {
			break
		}
		if pn, ok := ch.(*parse.Primary); ok && pn.Type == parse.Lambda {
			if pos >= pn.Range().To {
				continue
			}
		}
		eachDefinedVariable(ch, pos, f)
	}
}

// Calls f for each variable defined in fn.
func eachDefinedVariableInForm(fn *parse.Form, f func(string)) {
	if fn.Head == nil {
		return
	}
	switch head, _ := cmpd.StringLiteral(fn.Head); head {
	case "var":
		for _, arg := range fn.Args {
			if parse.SourceText(arg) == "=" {
				break
			}
			// TODO: This simplified version may not match the actual
			// algorithm used by the compiler to parse an LHS.
			if varRef, ok := cmpd.StringLiteral(arg); ok {
				_, name := eval.SplitSigil(varRef)
				f(name)
			}
		}
	case "fn":
		if len(fn.Args) >= 1 {
			if name, ok := cmpd.StringLiteral(fn.Args[0]); ok {
				f(name + eval.FnSuffix)
			}
		}
	}
}

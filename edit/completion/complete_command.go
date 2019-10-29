package completion

import (
	"strings"

	"github.com/elves/elvish/edit/ui"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type commandComplContext struct {
	complContextCommon
}

const quotingForEmptySeed = parse.Bareword

func findCommandComplContext(n parse.Node, ev pureEvaler) complContext {
	// Determine if we are starting a new command. There are 3 cases:
	// 1. The whole chunk is empty (nothing entered at all): the leaf is a
	//    Chunk.
	// 2. Just after a newline or semicolon: the leaf is a Sep and its parent is
	//    a Chunk.
	// 3. Just after a pipe: the leaf is a Sep and its parent is a Pipeline.
	if is(n, aChunk) {
		return &commandComplContext{
			complContextCommon{"", parse.Bareword, n.Range().To, n.Range().To}}
	}
	if is(n, aSep) {
		parent := n.Parent()
		switch {
		case is(parent, aChunk), is(parent, aPipeline):
			return &commandComplContext{
				complContextCommon{"", quotingForEmptySeed, n.Range().To, n.Range().To}}
		case is(parent, aPrimary):
			ptype := parent.(*parse.Primary).Type
			if ptype == parse.OutputCapture || ptype == parse.ExceptionCapture {
				return &commandComplContext{
					complContextCommon{"", quotingForEmptySeed, n.Range().To, n.Range().To}}
			}
		}
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return &commandComplContext{
						complContextCommon{seed, primary.Type, compound.Range().From, compound.Range().To}}
				}
			}
		}
	}
	return nil
}

func (*commandComplContext) name() string { return "command" }

func (ctx *commandComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	return complFormHeadInner(ctx.seed, env.evaler, ch)
}

func complFormHeadInner(head string, ev *eval.Evaler, rawCands chan<- rawCandidate) error {
	if util.DontSearch(head) {
		return complFilenameInner(head, true, rawCands)
	}

	got := func(s string) {
		rawCands <- plainCandidate(s)
	}
	for special := range eval.IsBuiltinSpecial {
		got(special)
	}
	sigil, qname := eval.SplitVariableRef(head)
	ns, _ := eval.SplitQNameNsIncomplete(qname)
	if sigil == "" {
		logger.Printf("completing commands in ns %q", ns)
		ev.EachVariableInTop(ns, func(varname string) {
			switch {
			case strings.HasSuffix(varname, eval.FnSuffix):
				got(ns + varname[:len(varname)-len(eval.FnSuffix)])
			case strings.HasSuffix(varname, eval.NsSuffix):
				got(ns + varname)
			default:
				name := ns + varname
				rawCands <- &complexCandidate{name, " = ", " = ", ui.Styles{}}
			}
		})
	}
	eval.EachExternal(func(command string) {
		got(command)
		if strings.HasPrefix(head, "e:") {
			got("e:" + command)
		}
	})
	return nil
}

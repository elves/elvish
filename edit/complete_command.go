package edit

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type commandCompleter struct {
	seed       string
	quoting    parse.PrimaryType
	begin, end int
}

const quotingForEmptySeed = parse.Bareword

func findCommandCompleter(n parse.Node, ev *eval.Evaler) completerIface {
	// Determine if we are starting a new command. There are 3 cases:
	// 1. The whole chunk is empty (nothing entered at all): the leaf is a
	//    Chunk.
	// 2. Just after a newline or semicolon: the leaf is a Sep and its parent is
	//    a Chunk.
	// 3. Just after a pipe: the leaf is a Sep and its parent is a Pipeline.
	if parse.IsChunk(n) {
		return &commandCompleter{"", parse.Bareword, n.End(), n.End()}
	}
	if parse.IsSep(n) {
		parent := n.Parent()
		switch {
		case parse.IsChunk(parent), parse.IsPipeline(parent):
			return &commandCompleter{"", quotingForEmptySeed, n.End(), n.End()}
		case parse.IsPrimary(parent):
			ptype := parent.(*parse.Primary).Type
			if ptype == parse.OutputCapture || ptype == parse.ExceptionCapture {
				return &commandCompleter{"", quotingForEmptySeed, n.End(), n.End()}
			}
		}
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return &commandCompleter{seed, primary.Type, compound.Begin(), compound.End()}
				}
			}
		}
	}
	return nil
}

func (*commandCompleter) name() string { return "command" }

func (compl *commandCompleter) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = complFormHeadInner(compl.seed, ev, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, compl.seed, rawCands, compl.quoting)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}
	return &complSpec{compl.begin, compl.end, cands}, nil
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
	explode, ns, _ := eval.ParseVariable(head)
	if !explode {
		iterateVariables(ev, ns, func(varname string) {
			if strings.HasPrefix(varname, eval.FnPrefix) {
				got(eval.MakeVariableName(false, ns, varname[len(eval.FnPrefix):]))
			} else {
				got(eval.MakeVariableName(false, ns, varname) + "=")
			}
		})
	}
	ev.EachExternal(func(command string) {
		got(command)
		if strings.HasPrefix(head, "e:") {
			got("e:" + command)
		}
	})
	// TODO Support non-module namespaces.
	for ns := range ev.Global.Uses {
		if head != ns+":" {
			got(ns + ":")
		}
	}
	for ns := range ev.Builtin.Uses {
		if head != ns+":" {
			got(ns + ":")
		}
	}
	return nil
}

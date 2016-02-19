package eval

import (
	"errors"

	"github.com/elves/elvish/parse"
)

// VariablesOp is an operation on an EvalCtx that produce Variable's.
type VariablesOp struct {
	Func       VariablesOpFunc
	Begin, End int
}

type VariablesOpFunc func(*EvalCtx) []Variable

func (op VariablesOp) Exec(ec *EvalCtx) []Variable {
	ec.begin, ec.end = op.Begin, op.End
	return op.Func(ec)
}

func (cp *compiler) multiVariable(n *parse.Indexing) VariablesOpFunc {
	if n.Head.Type == parse.Braced {
		// XXX ignore n.Indicies.
		compounds := n.Head.Braced
		indexings := make([]*parse.Indexing, len(compounds))
		for i, cn := range compounds {
			if len(cn.Indexings) != 1 {
				cp.compiling(cn)
				cp.errorf("must be a variable spec")
			}
			indexings[i] = cn.Indexings[0]
		}
		ops := cp.singleVariableOps(indexings, "must be a variable spec")
		return func(ec *EvalCtx) []Variable {
			var variables []Variable
			for _, op := range ops {
				variables = append(variables, op.Exec(ec)...)
			}
			return variables
		}
	}
	return cp.singleVariable(n, "must be a variable spec or a braced list of those")
}

func (cp *compiler) singleVariable(n *parse.Indexing, msg string) VariablesOpFunc {
	// XXX will we be using this for purposes other than setting?
	varname := cp.literal(n.Head, msg)

	if len(n.Indicies) == 0 {
		cp.registerVariableSet(varname)

		return func(ec *EvalCtx) []Variable {
			splice, ns, barename := parseVariable(varname)
			if splice {
				// XXX
				ec.errorf("not yet supported")
			}
			variable := ec.ResolveVar(ns, barename)
			if variable == nil {
				if ns == "" || ns == "local" {
					// New variable.
					// XXX We depend on the fact that this variable will
					// immeidately be set.
					variable = NewPtrVariable(nil)
					ec.local[barename] = variable
				} else if mod, ok := ec.modules[ns]; ok {
					variable = NewPtrVariable(nil)
					mod[barename] = variable
				} else {
					ec.errorf("cannot set $%s", varname)
				}
			}
			return []Variable{variable}
		}
	}
	cp.registerVariableGet(varname)
	indexOps := cp.arrayOps(n.Indicies)

	p := n.Begin()
	indexBegins := make([]int, len(n.Indicies))
	indexEnds := make([]int, len(n.Indicies))
	for i, in := range n.Indicies {
		indexBegins[i] = in.Begin()
		indexEnds[i] = in.End()
	}

	return func(ec *EvalCtx) []Variable {
		splice, ns, name := parseVariable(varname)
		if splice {
			// XXX
			ec.errorf("not yet supported")
		}
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			ec.errorf("variable $%s does not exisit, compiler bug", varname)
		}
		if len(indexOps) == 0 {
			// Just a variable, return directly.
			return []Variable{variable}
		}

		// Indexing. Do Index up to the last but one index.
		value := variable.Get()
		n := len(indexOps)
		// TODO set location information according.
		for _, op := range indexOps[:n-1] {
			indexer := mustIndexer(value, ec)

			indicies := op.Exec(ec)
			values := indexer.Index(indicies)
			if len(values) != 1 {
				throw(errors.New("multi indexing not implemented"))
			}
			value = values[0]
		}
		// Now this must be an IndexSetter.
		indexSetter, ok := value.(IndexSetter)
		if !ok {
			// XXX the indicated end location will fall on or after the opening
			// bracket of the last index, instead of exactly on the penultimate
			// index.
			ec.errorpf(p, indexBegins[n-1], "cannot be indexed for setting (value is %s, type %s)", value.Repr(NoPretty), value.Kind())
		}
		// XXX Duplicate code.
		indicies := indexOps[n-1].Exec(ec)
		if len(indicies) != 1 {
			ec.errorpf(indexBegins[n-1], indexEnds[n-1], "index must eval to a single Value (got %v)", indicies)
		}
		return []Variable{elemVariable{indexSetter, indicies[0]}}
	}
}

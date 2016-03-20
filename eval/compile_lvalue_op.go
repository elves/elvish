package eval

import (
	"errors"

	"github.com/elves/elvish/parse"
)

// VariablesOp is an operation on an EvalCtx that produce Variable's.
type LValuesOp struct {
	Func       LValuesOpFunc
	Begin, End int
}

type LValuesOpFunc func(*EvalCtx) []Variable

func (op LValuesOp) Exec(ec *EvalCtx) []Variable {
	// Empty value is considered to generate no lvalues.
	if op.Func == nil {
		return []Variable{}
	}
	ec.begin, ec.end = op.Begin, op.End
	return op.Func(ec)
}

// lvaluesOp compiles lvalues, returning the fixed part and, optionally a rest
// part.
//
// In the AST an lvalue is either an Indexing node where the head is a string
// literal, or a braced list of such Indexing nodes. The last Indexing node may
// be prefixed by @, in which case they become the rest part. For instance, in
// {a[x],b,@c[z]}, "a[x],b" is the fixed part and "c[z]" is the rest part.
func (cp *compiler) lvaluesOp(n *parse.Indexing) (LValuesOp, LValuesOp) {
	if n.Head.Type == parse.Braced {
		// Braced list of variable specs, possibly with indicies. The braced list
		if len(n.Indicies) > 0 {
			cp.errorf("may not have indicies")
		}

		opFuncs := make([]LValuesOpFunc, len(n.Head.Braced))
		var restNode *parse.Indexing
		var restOpFunc LValuesOpFunc

		// Compile each spec inside the brace.
		lvalueNodes := n.Head.Braced
		fixedEnd := 0
		for i, cn := range lvalueNodes {
			if len(cn.Indexings) != 1 {
				cp.errorpf(cn.Begin(), cn.End(), "must be an lvalue")
			}
			var rest bool
			rest, opFuncs[i] = cp.lvaluesOne(cn.Indexings[0], "must be an lvalue ")
			// Only the last one may a rest part.
			if rest {
				if i == len(n.Head.Braced)-1 {
					restNode = cn.Indexings[0]
					restOpFunc = opFuncs[i]
				} else {
					cp.errorpf(cn.Begin(), cn.End(), "only the last lvalue may have @")
				}
			} else {
				fixedEnd = cn.End()
			}
		}

		var restOp LValuesOp
		// If there is a rest part, make LValuesOp for it and remove it from opFuncs.
		if restOpFunc != nil {
			restOp = LValuesOp{restOpFunc, restNode.Begin(), restNode.End()}
			opFuncs = opFuncs[:len(opFuncs)-1]
		}

		var op LValuesOp
		// If there is still anything left in opFuncs, make LValuesOp for the fixed part.
		if len(opFuncs) > 0 {
			op = LValuesOp{func(ec *EvalCtx) []Variable {
				var variables []Variable
				for _, opFunc := range opFuncs {
					variables = append(variables, opFunc(ec)...)
				}
				return variables
			}, lvalueNodes[0].Begin(), fixedEnd}
		}

		return op, restOp
	}
	rest, opFunc := cp.lvaluesOne(n, "must be an lvalue or a braced list of those")
	op := LValuesOp{opFunc, n.Begin(), n.End()}
	if rest {
		return LValuesOp{}, op
	} else {
		return op, LValuesOp{}
	}
}

func (cp *compiler) lvaluesOne(n *parse.Indexing, msg string) (bool, LValuesOpFunc) {
	varname := cp.literal(n.Head, msg)
	cp.registerVariableSet(varname)
	splice, ns, barename := ParseVariable(varname)

	if len(n.Indicies) == 0 {
		return splice, func(ec *EvalCtx) []Variable {
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

	p := n.Begin()
	indexOps := cp.arrayOps(n.Indicies)

	return splice, func(ec *EvalCtx) []Variable {
		variable := ec.ResolveVar(ns, barename)
		if variable == nil {
			ec.errorf("variable $%s does not exisit, compiler bug", varname)
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
			ec.errorpf(p, indexOps[n-1].Begin, "cannot be indexed for setting (value is %s, type %s)", value.Repr(NoPretty), value.Kind())
		}
		// XXX Duplicate code.
		indicies := indexOps[n-1].Exec(ec)
		if len(indicies) != 1 {
			ec.errorpf(indexOps[n-1].Begin, indexOps[n-1].End, "index must eval to a single Value (got %v)", indicies)
		}
		return []Variable{elemVariable{indexSetter, indicies[0]}}
	}
}

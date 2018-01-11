package eval

import (
	"errors"
	"strings"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
)

// LValuesOp is an operation on an EvalCtx that produce Variable's.
type LValuesOp struct {
	Func       LValuesOpFunc
	Begin, End int
}

// LValuesOpFunc is the body of an LValuesOp.
type LValuesOpFunc func(*Frame) []vartypes.Variable

// Exec executes an LValuesOp, producing Variable's.
func (op LValuesOp) Exec(ec *Frame) []vartypes.Variable {
	// Empty value is considered to generate no lvalues.
	if op.Func == nil {
		return []vartypes.Variable{}
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
		// Braced list of variable specs, possibly with indicies.
		if len(n.Indicies) > 0 {
			cp.errorf("may not have indicies")
		}
		return cp.lvaluesMulti(n.Head.Braced)
	}
	rest, opFunc := cp.lvalueBase(n, "must be an lvalue or a braced list of those")
	op := LValuesOp{opFunc, n.Begin(), n.End()}
	if rest {
		return LValuesOp{}, op
	}
	return op, LValuesOp{}
}

func (cp *compiler) lvaluesMulti(nodes []*parse.Compound) (LValuesOp, LValuesOp) {
	opFuncs := make([]LValuesOpFunc, len(nodes))
	var restNode *parse.Indexing
	var restOpFunc LValuesOpFunc

	// Compile each spec inside the brace.
	fixedEnd := 0
	for i, cn := range nodes {
		if len(cn.Indexings) != 1 {
			cp.errorpf(cn.Begin(), cn.End(), "must be an lvalue")
		}
		var rest bool
		rest, opFuncs[i] = cp.lvalueBase(cn.Indexings[0], "must be an lvalue ")
		// Only the last one may a rest part.
		if rest {
			if i == len(nodes)-1 {
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
		op = LValuesOp{func(ec *Frame) []vartypes.Variable {
			var variables []vartypes.Variable
			for _, opFunc := range opFuncs {
				variables = append(variables, opFunc(ec)...)
			}
			return variables
		}, nodes[0].Begin(), fixedEnd}
	}

	return op, restOp
}

func (cp *compiler) lvalueBase(n *parse.Indexing, msg string) (bool, LValuesOpFunc) {
	qname := cp.literal(n.Head, msg)
	explode, ns, name := ParseVariable(qname)
	if len(n.Indicies) == 0 {
		return explode, cp.lvalueVariable(ns, name)
	}
	return explode, cp.lvalueElement(ns, name, n)
}

func (cp *compiler) lvalueVariable(ns, name string) LValuesOpFunc {
	cp.registerVariableSet(ns, name)

	return func(ec *Frame) []vartypes.Variable {
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			if ns == "" || ns == "local" {
				// New variable.
				// XXX We depend on the fact that this variable will
				// immeidately be set.
				if strings.HasSuffix(name, FnSuffix) {
					variable = vartypes.NewValidatedPtr(nil, ShouldBeFn)
				} else if strings.HasSuffix(name, NsSuffix) {
					variable = vartypes.NewValidatedPtr(nil, ShouldBeNs)
				} else {
					variable = vartypes.NewPtr(nil)
				}
				ec.local[name] = variable
			} else {
				throwf("new variables can only be created in local scope")
			}
		}
		return []vartypes.Variable{variable}
	}
}

func (cp *compiler) lvalueElement(ns, name string, n *parse.Indexing) LValuesOpFunc {
	begin, end := n.Begin(), n.End()
	ends := make([]int, len(n.Indicies)+1)
	ends[0] = n.Head.End()
	for i, idx := range n.Indicies {
		ends[i+1] = idx.End()
	}

	indexOps := cp.arrayOps(n.Indicies)

	return func(ec *Frame) []vartypes.Variable {
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			throwf("variable $%s:%s does not exist, compiler bug", ns, name)
		}

		indicies := make([]types.Value, len(indexOps))
		for i, op := range indexOps {
			values := op.Exec(ec)
			// TODO: Implement multi-indexing.
			if len(values) != 1 {
				throw(errors.New("multi indexing not implemented"))
			}
			indicies[i] = values[0]
		}
		elemVar, err := vartypes.MakeElement(variable, indicies)
		if err != nil {
			level := vartypes.GetMakeElementErrorLevel(err)
			if level < 0 {
				ec.errorpf(begin, end, "%s", err)
			} else {
				ec.errorpf(begin, ends[level], "%s", err)
			}
		}
		return []vartypes.Variable{elemVar}
	}
}

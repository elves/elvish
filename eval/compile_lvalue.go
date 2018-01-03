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
		// Braced list of variable specs, possibly with indicies. The braced list
		if len(n.Indicies) > 0 {
			cp.errorf("may not have indicies")
		}
		return cp.lvaluesMulti(n.Head.Braced)
	}
	rest, opFunc := cp.lvaluesOne(n, "must be an lvalue or a braced list of those")
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
		rest, opFuncs[i] = cp.lvaluesOne(cn.Indexings[0], "must be an lvalue ")
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

func (cp *compiler) lvaluesOne(n *parse.Indexing, msg string) (bool, LValuesOpFunc) {
	varname := cp.literal(n.Head, msg)
	cp.registerVariableSetQname(varname)
	explode, ns, barename := ParseVariable(varname)

	if len(n.Indicies) == 0 {
		return explode, func(ec *Frame) []vartypes.Variable {
			variable := ec.ResolveVar(ns, barename)
			if variable == nil {
				if ns == "" || ns == "local" {
					// New variable.
					// XXX We depend on the fact that this variable will
					// immeidately be set.
					if strings.HasSuffix(barename, FnSuffix) {
						variable = vartypes.NewValidatedPtrVariable(nil, ShouldBeFn)
					} else if strings.HasSuffix(barename, NsSuffix) {
						variable = vartypes.NewValidatedPtrVariable(nil, ShouldBeNs)
					} else {
						variable = vartypes.NewPtrVariable(nil)
					}
					ec.local[barename] = variable
				} else if mod, ok := ec.modules[ns]; ok {
					variable = vartypes.NewPtrVariable(nil)
					mod[barename] = variable
				} else {
					throwf("cannot set $%s", varname)
				}
			}
			return []vartypes.Variable{variable}
		}
	}

	headBegin, headEnd := n.Head.Begin(), n.Head.End()
	indexOps := cp.arrayOps(n.Indicies)

	return explode, func(ec *Frame) []vartypes.Variable {
		variable := ec.ResolveVar(ns, barename)
		if variable == nil {
			throwf("variable $%s does not exist, compiler bug", varname)
		}

		// Evaluate assocers and indices.
		// Assignment of indexed variables actually assignes the variable, with
		// the right hand being a nested series of Assocs. As the simplest
		// example, `a[0] = x` is equivalent to `a = (assoc $a 0 x)`. A more
		// complex example is that `a[0][1][2] = x` is equivalent to
		//	`a = (assoc $a 0 (assoc $a[0] 1 (assoc $a[0][1] 2 x)))`.
		// Note that in each assoc form, the first two arguments can be
		// determined now, while the last argument is only known when the
		// right-hand-side is known. So here we evaluate the first two arguments
		// of each assoc form and put them in two slices, assocers and indicies.
		// In the previous example, the two slices will contain:
		//
		// assocers: $a $a[0] $a[0][1]
		// indicies:  0     1        2
		//
		// When the right-hand side of the assignment becomes available, the new
		// value for $a is evaluated by doing Assoc from inside out.
		assocers := make([]types.Assocer, len(indexOps))
		indicies := make([]types.Value, len(indexOps))
		varValue, ok := variable.Get().(IndexOneAssocer)
		if !ok {
			ec.errorpf(headBegin, headEnd, "cannot be indexed for setting")
		}
		assocers[0] = varValue
		for i, op := range indexOps {
			var lastAssocer types.IndexOneer
			if i < len(indexOps)-1 {
				var ok bool
				lastAssocer, ok = assocers[i].(types.IndexOneer)
				if !ok {
					// This cannot occur when i==0, since varValue as already
					// asserted to be an IndexOnner.
					ec.errorpf(headBegin, indexOps[i-1].End, "cannot be indexed")
				}
			}

			values := op.Exec(ec)
			// TODO: Implement multi-indexing.
			if len(values) != 1 {
				throw(errors.New("multi indexing not implemented"))
			}
			index := values[0]
			indicies[i] = index

			if i < len(indexOps)-1 {
				assocer, ok := lastAssocer.IndexOne(index).(types.Assocer)
				if !ok {
					ec.errorpf(headBegin, indexOps[i].End,
						"cannot be indexed for setting")
				}
				assocers[i+1] = assocer
			}
		}
		return []vartypes.Variable{&elemVariable{variable, assocers, indicies, nil}}
	}
}

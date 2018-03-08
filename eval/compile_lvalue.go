package eval

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
)

// LValuesOp is an operation on an Frame that produce Variable's.
type LValuesOp struct {
	Body       LValuesOpBody
	Begin, End int
}

// LValuesOpBody is the body of an LValuesOp.
type LValuesOpBody interface {
	Invoke(*Frame) ([]vars.Var, error)
}

// Exec executes an LValuesOp, producing Variable's.
func (op LValuesOp) Exec(fm *Frame) ([]vars.Var, error) {
	// Empty value is considered to generate no lvalues.
	if op.Body == nil {
		return []vars.Var{}, nil
	}
	fm.begin, fm.end = op.Begin, op.End
	return op.Body.Invoke(fm)
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
	opFuncs := make([]LValuesOpBody, len(nodes))
	var restNode *parse.Indexing
	var restOpFunc LValuesOpBody

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
		op = LValuesOp{seqLValuesOpBody{opFuncs}, nodes[0].Begin(), fixedEnd}
	}

	return op, restOp
}

func (cp *compiler) lvalueBase(n *parse.Indexing, msg string) (bool, LValuesOpBody) {
	qname := cp.literal(n.Head, msg)
	explode, ns, name := ParseVariableRef(qname)
	if len(n.Indicies) == 0 {
		cp.registerVariableSet(ns, name)
		return explode, varOp{ns, name}
	}
	return explode, cp.lvalueElement(ns, name, n)
}

func (cp *compiler) lvalueElement(ns, name string, n *parse.Indexing) LValuesOpBody {
	cp.registerVariableGet(ns, name)

	begin, end := n.Begin(), n.End()
	ends := make([]int, len(n.Indicies)+1)
	ends[0] = n.Head.End()
	for i, idx := range n.Indicies {
		ends[i+1] = idx.End()
	}

	indexOps := cp.arrayOps(n.Indicies)

	return &elemOp{ns, name, indexOps, begin, end, ends}
}

type seqLValuesOpBody struct {
	ops []LValuesOpBody
}

func (op seqLValuesOpBody) Invoke(fm *Frame) ([]vars.Var, error) {
	var variables []vars.Var
	for _, op := range op.ops {
		moreVariables, err := op.Invoke(fm)
		if err != nil {
			return nil, err
		}
		variables = append(variables, moreVariables...)
	}
	return variables, nil
}

type varOp struct {
	ns, name string
}

func (op varOp) Invoke(fm *Frame) ([]vars.Var, error) {
	variable := fm.ResolveVar(op.ns, op.name)
	if variable == nil {
		if op.ns == "" || op.ns == "local" {
			// New variable.
			// XXX We depend on the fact that this variable will
			// immeidately be set.
			if strings.HasSuffix(op.name, FnSuffix) {
				val := Callable(nil)
				variable = vars.FromPtr(&val)
			} else if strings.HasSuffix(op.name, NsSuffix) {
				val := Ns(nil)
				variable = vars.FromPtr(&val)
			} else {
				variable = vars.NewAnyWithInit(nil)
			}
			fm.local[op.name] = variable
		} else {
			return nil, fmt.Errorf("new variables can only be created in local scope")
		}
	}
	return []vars.Var{variable}, nil
}

type elemOp struct {
	ns       string
	name     string
	indexOps []ValuesOp
	begin    int
	end      int
	ends     []int
}

func (op *elemOp) Invoke(fm *Frame) ([]vars.Var, error) {
	variable := fm.ResolveVar(op.ns, op.name)
	if variable == nil {
		return nil, fmt.Errorf("variable $%s:%s does not exist, compiler bug", op.ns, op.name)
	}

	indicies := make([]interface{}, len(op.indexOps))
	for i, op := range op.indexOps {
		values, err := op.Exec(fm)
		if err != nil {
			return nil, err
		}
		// TODO: Implement multi-indexing.
		if len(values) != 1 {
			return nil, errors.New("multi indexing not implemented")
		}
		indicies[i] = values[0]
	}
	elemVar, err := vars.MakeElement(variable, indicies)
	if err != nil {
		level := vars.ElementErrorLevel(err)
		if level < 0 {
			fm.errorpf(op.begin, op.end, "%s", err)
		} else {
			fm.errorpf(op.begin, op.ends[level], "%s", err)
		}
	}
	return []vars.Var{elemVar}, nil
}

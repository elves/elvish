package eval

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
)

// lvaluesOp compiles lvalues, returning the fixed part and, optionally a rest
// part.
//
// In the AST an lvalue is either an Indexing node where the head is a string
// literal, or a braced list of such Indexing nodes. The last Indexing node may
// be prefixed by @, in which case they become the rest part. For instance, in
// {a[x],b,@c[z]}, "a[x],b" is the fixed part and "c[z]" is the rest part.
func (cp *compiler) lvaluesOp(n *parse.Indexing) (lvaluesOp, lvaluesOp) {
	if n.Head.Type == parse.Braced {
		// Braced list of variable specs, possibly with indicies.
		if len(n.Indicies) > 0 {
			cp.errorf("may not have indicies")
		}
		return cp.lvaluesMulti(n.Head.Braced)
	}
	rest, opFunc := cp.lvalueBase(n, "must be an lvalue or a braced list of those")
	op := lvaluesOp{opFunc, n.Range().From, n.Range().To}
	if rest {
		return lvaluesOp{}, op
	}
	return op, lvaluesOp{}
}

func (cp *compiler) lvaluesMulti(nodes []*parse.Compound) (lvaluesOp, lvaluesOp) {
	opFuncs := make([]lvaluesOpBody, len(nodes))
	var restNode *parse.Indexing
	var restOpFunc lvaluesOpBody

	// Compile each spec inside the brace.
	fixedEnd := 0
	for i, cn := range nodes {
		if len(cn.Indexings) != 1 {
			cp.errorpf(cn.Range().From, cn.Range().To, "must be an lvalue")
		}
		var rest bool
		rest, opFuncs[i] = cp.lvalueBase(cn.Indexings[0], "must be an lvalue ")
		// Only the last one may a rest part.
		if rest {
			if i == len(nodes)-1 {
				restNode = cn.Indexings[0]
				restOpFunc = opFuncs[i]
			} else {
				cp.errorpf(cn.Range().From, cn.Range().To, "only the last lvalue may have @")
			}
		} else {
			fixedEnd = cn.Range().To
		}
	}

	var restOp lvaluesOp
	// If there is a rest part, make LValuesOp for it and remove it from opFuncs.
	if restOpFunc != nil {
		restOp = lvaluesOp{restOpFunc, restNode.Range().From, restNode.Range().To}
		opFuncs = opFuncs[:len(opFuncs)-1]
	}

	var op lvaluesOp
	// If there is still anything left in opFuncs, make LValuesOp for the fixed part.
	if len(opFuncs) > 0 {
		op = lvaluesOp{seqLValuesOpBody{opFuncs}, nodes[0].Range().From, fixedEnd}
	}

	return op, restOp
}

func (cp *compiler) lvalueBase(n *parse.Indexing, msg string) (bool, lvaluesOpBody) {
	ref := cp.literal(n.Head, msg)
	sigil, qname := SplitVariableRef(ref)
	// TODO: Deal with other sigils too
	explode := sigil != ""
	if len(n.Indicies) == 0 {
		cp.registerVariableSet(qname)
		return explode, varOp{qname}
	}
	return explode, cp.lvalueElement(qname, n)
}

func (cp *compiler) lvalueElement(qname string, n *parse.Indexing) lvaluesOpBody {
	cp.registerVariableGet(qname)

	begin, end := n.Range().From, n.Range().To
	ends := make([]int, len(n.Indicies)+1)
	ends[0] = n.Head.Range().To
	for i, idx := range n.Indicies {
		ends[i+1] = idx.Range().To
	}

	indexOps := cp.arrayOps(n.Indicies)

	return &elemOp{qname, indexOps, begin, end, ends}
}

type seqLValuesOpBody struct {
	ops []lvaluesOpBody
}

func (op seqLValuesOpBody) invoke(fm *Frame) ([]vars.Var, error) {
	var variables []vars.Var
	for _, op := range op.ops {
		moreVariables, err := op.invoke(fm)
		if err != nil {
			return nil, err
		}
		variables = append(variables, moreVariables...)
	}
	return variables, nil
}

type varOp struct {
	qname string
}

func (op varOp) invoke(fm *Frame) ([]vars.Var, error) {
	variable := fm.ResolveVar(op.qname)
	if variable == nil {
		ns, name := SplitQNameNs(op.qname)
		if ns == "" || ns == ":" || ns == "local:" {
			// New variable.
			// XXX We depend on the fact that this variable will
			// immeidately be set.
			if strings.HasSuffix(name, FnSuffix) {
				val := Callable(nil)
				variable = vars.FromPtr(&val)
			} else if strings.HasSuffix(name, NsSuffix) {
				val := Ns(nil)
				variable = vars.FromPtr(&val)
			} else {
				variable = vars.FromInit(nil)
			}
			fm.local[name] = variable
		} else {
			return nil, fmt.Errorf("new variables can only be created in local scope")
		}
	}
	return []vars.Var{variable}, nil
}

type elemOp struct {
	qname    string
	indexOps []valuesOp
	begin    int
	end      int
	ends     []int
}

func (op *elemOp) invoke(fm *Frame) ([]vars.Var, error) {
	variable := fm.ResolveVar(op.qname)
	if variable == nil {
		return nil, fmt.Errorf("variable $%s does not exist, compiler bug", op.qname)
	}

	indicies := make([]interface{}, len(op.indexOps))
	for i, op := range op.indexOps {
		values, err := op.exec(fm)
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
			return nil, fm.errorpf(op.begin, op.end, "%s", err)
		}
		return nil, fm.errorpf(op.begin, op.ends[level], "%s", err)
	}
	return []vars.Var{elemVar}, nil
}

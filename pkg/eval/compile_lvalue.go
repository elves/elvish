package eval

import (
	"errors"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

// Parsed group of lvalues.
type lvaluesGroup struct {
	lvalues []lvalue
	// Index of the rest variable within lvalues. If there is no rest variable,
	// the index is -1.
	rest int
}

// Parsed lvalue.
type lvalue struct {
	diag.Ranging
	ref      *varRef
	indexOps []valuesOp
	ends     []int
}

func (cp *compiler) parseCompoundLValues(ns []*parse.Compound) lvaluesGroup {
	g := lvaluesGroup{nil, -1}
	for _, n := range ns {
		if len(n.Indexings) != 1 {
			cp.errorpf(n, "lvalue may not be composite expressions")
		}
		more := cp.parseIndexingLValue(n.Indexings[0])
		if more.rest == -1 {
			g.lvalues = append(g.lvalues, more.lvalues...)
		} else if g.rest != -1 {
			cp.errorpf(n, "at most one rest variable is allowed")
		} else {
			g.rest = len(g.lvalues) + more.rest
			g.lvalues = append(g.lvalues, more.lvalues...)
		}
	}
	return g
}

func (cp *compiler) parseIndexingLValue(n *parse.Indexing) lvaluesGroup {
	if n.Head.Type == parse.Braced {
		// Braced list of lvalues may not have indices.
		if len(n.Indicies) > 0 {
			cp.errorpf(n, "braced list may not have indices when used as lvalue")
		}
		return cp.parseCompoundLValues(n.Head.Braced)
	}
	// A basic lvalue.
	if !parse.ValidLHSVariable(n.Head, true) {
		cp.errorpf(n.Head, "lvalue must be valid literal variable names")
	}
	varUse := n.Head.Value
	sigil, qname := SplitSigil(varUse)
	var ref *varRef
	if len(n.Indicies) == 0 {
		ref = resolveVarRef(cp, qname, nil)
		if ref == nil {
			segs := SplitQNameSegs(qname)
			if len(segs) == 1 {
				// Unqualified name - implicit local
				ref = &varRef{localScope, cp.thisScope().addInner(segs[0]), nil}
			} else if len(segs) == 2 && (segs[0] == "local:" || segs[0] == ":") {
				// Qualified local name
				ref = &varRef{localScope, cp.thisScope().addInner(segs[1]), nil}
			} else {
				cp.errorpf(n, "cannot create variable $%s; new variables can only be created in the local scope", qname)
			}
		}
	} else {
		ref = resolveVarRef(cp, qname, n)
		if ref == nil {
			cp.errorpf(n, "cannot find variable $%s", qname)
		}
	}
	ends := make([]int, len(n.Indicies)+1)
	ends[0] = n.Head.Range().To
	for i, idx := range n.Indicies {
		ends[i+1] = idx.Range().To
	}
	lv := lvalue{n.Range(), ref, cp.arrayOps(n.Indicies), ends}
	restIndex := -1
	if sigil == "@" {
		restIndex = 0
	}
	// TODO: Deal with other sigils when they exist.
	return lvaluesGroup{[]lvalue{lv}, restIndex}
}

type assignOp struct {
	diag.Ranging
	lhs lvaluesGroup
	rhs valuesOp
}

func (op *assignOp) exec(fm *Frame) Exception {
	variables := make([]vars.Var, len(op.lhs.lvalues))
	for i, lvalue := range op.lhs.lvalues {
		variable, err := derefLValue(fm, lvalue)
		if err != nil {
			return fm.errorp(op, err)
		}
		variables[i] = variable
	}

	values, exc := op.rhs.exec(fm)
	if exc != nil {
		return exc
	}

	if op.lhs.rest == -1 {
		if len(variables) != len(values) {
			return fm.errorp(op, errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: len(variables), ValidHigh: len(variables), Actual: len(values)})
		}
		for i, variable := range variables {
			err := variable.Set(values[i])
			if err != nil {
				return fm.errorp(op, err)
			}
		}
	} else {
		if len(values) < len(variables)-1 {
			return fm.errorp(op, errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: len(variables) - 1, ValidHigh: -1, Actual: len(values)})
		}
		rest := op.lhs.rest
		for i := 0; i < rest; i++ {
			err := variables[i].Set(values[i])
			if err != nil {
				return fm.errorp(op, err)
			}
		}
		restOff := len(values) - len(variables)
		err := variables[rest].Set(vals.MakeList(values[rest : rest+restOff+1]...))
		if err != nil {
			return fm.errorp(op, err)
		}
		for i := rest + 1; i < len(variables); i++ {
			err := variables[i].Set(values[i+restOff])
			if err != nil {
				return fm.errorp(op, err)
			}
		}
	}
	return nil
}

// NoSuchVariable returns an error representing that a variable can't be found.
func NoSuchVariable(name string) error {
	return noSuchVariableError{name}
}

type noSuchVariableError struct{ name string }

func (er noSuchVariableError) Error() string { return "no variable $" + er.name }

func derefLValue(fm *Frame, lv lvalue) (vars.Var, error) {
	variable := deref(fm, lv.ref)
	if variable == nil {
		return nil, NoSuchVariable(fm.srcMeta.Code[lv.From:lv.To])
	}
	if len(lv.indexOps) == 0 {
		return variable, nil
	}
	indices := make([]interface{}, len(lv.indexOps))
	for i, op := range lv.indexOps {
		values, exc := op.exec(fm)
		if exc != nil {
			return nil, exc
		}
		// TODO: Implement multi-indexing.
		if len(values) != 1 {
			return nil, errors.New("multi indexing not implemented")
		}
		indices[i] = values[0]
	}
	elemVar, err := vars.MakeElement(variable, indices)
	if err != nil {
		level := vars.ElementErrorLevel(err)
		if level < 0 {
			return nil, fm.errorp(lv, err)
		}
		return nil, fm.errorp(diag.Ranging{From: lv.From, To: lv.ends[level]}, err)
	}
	return elemVar, nil
}

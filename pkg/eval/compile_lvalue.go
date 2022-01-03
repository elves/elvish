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

type lvalueFlag uint

const (
	setLValue lvalueFlag = 1 << iota
	newLValue
)

func (cp *compiler) parseCompoundLValues(ns []*parse.Compound, f lvalueFlag) lvaluesGroup {
	g := lvaluesGroup{nil, -1}
	for _, n := range ns {
		if len(n.Indexings) != 1 {
			cp.errorpf(n, "lvalue may not be composite expressions")
		}
		more := cp.parseIndexingLValue(n.Indexings[0], f)
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

func (cp *compiler) parseIndexingLValue(n *parse.Indexing, f lvalueFlag) lvaluesGroup {
	if n.Head.Type == parse.Braced {
		// Braced list of lvalues may not have indices.
		if len(n.Indices) > 0 {
			cp.errorpf(n, "braced list may not have indices when used as lvalue")
		}
		return cp.parseCompoundLValues(n.Head.Braced, f)
	}
	// A basic lvalue.
	if !parse.ValidLHSVariable(n.Head, true) {
		cp.errorpf(n.Head, "lvalue must be valid literal variable names")
	}
	varUse := n.Head.Value
	sigil, qname := SplitSigil(varUse)

	var ref *varRef
	if f&setLValue != 0 {
		ref = resolveVarRef(cp, qname, n)
		if ref != nil && len(ref.subNames) == 0 && ref.info.readOnly {
			cp.errorpf(n, "variable $%s is read-only", qname)
		}
	}
	if ref == nil {
		if f&newLValue == 0 {
			cp.errorpf(n, "cannot find variable $%s", qname)
		}
		if len(n.Indices) > 0 {
			cp.errorpf(n, "name for new variable must not have indices")
		}
		segs := SplitQNameSegs(qname)
		if len(segs) == 1 {
			// Unqualified name - implicit local
			name := segs[0]
			ref = &varRef{localScope,
				staticVarInfo{name, false, false}, cp.thisScope().add(name), nil}
		} else {
			cp.errorpf(n, "cannot create variable $%s; new variables can only be created in the local scope", qname)
		}
	}

	ends := make([]int, len(n.Indices)+1)
	ends[0] = n.Head.Range().To
	for i, idx := range n.Indices {
		ends[i+1] = idx.Range().To
	}
	lv := lvalue{n.Range(), ref, cp.arrayOps(n.Indices), ends}
	restIndex := -1
	if sigil == "@" {
		restIndex = 0
	}
	// TODO: Support % (and other sigils?) if https://b.elv.sh/584 is implemented for map explosion.
	return lvaluesGroup{[]lvalue{lv}, restIndex}
}

type assignOp struct {
	diag.Ranging
	lhs  lvaluesGroup
	rhs  valuesOp
	temp bool
}

func (op *assignOp) exec(fm *Frame) Exception {
	variables := make([]vars.Var, len(op.lhs.lvalues))
	for i, lvalue := range op.lhs.lvalues {
		variable, err := derefLValue(fm, lvalue)
		if err != nil {
			return fm.errorp(op.lhs.lvalues[i], err)
		}
		variables[i] = variable
	}

	values, exc := op.rhs.exec(fm)
	if exc != nil {
		return exc
	}

	rest, temp := op.lhs.rest, op.temp
	if rest == -1 {
		if len(variables) != len(values) {
			return fm.errorp(op, errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: len(variables), ValidHigh: len(variables), Actual: len(values)})
		}
		for i, variable := range variables {
			exc := set(fm, op.lhs.lvalues[i], temp, variable, values[i])
			if exc != nil {
				return exc
			}
		}
	} else {
		if len(values) < len(variables)-1 {
			return fm.errorp(op, errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: len(variables) - 1, ValidHigh: -1, Actual: len(values)})
		}
		for i := 0; i < rest; i++ {
			exc := set(fm, op.lhs.lvalues[i], temp, variables[i], values[i])
			if exc != nil {
				return exc
			}
		}
		restOff := len(values) - len(variables)
		exc := set(fm, op.lhs.lvalues[rest], temp,
			variables[rest], vals.MakeList(values[rest:rest+restOff+1]...))
		if exc != nil {
			return exc
		}
		for i := rest + 1; i < len(variables); i++ {
			exc := set(fm, op.lhs.lvalues[i], temp, variables[i], values[i+restOff])
			if exc != nil {
				return exc
			}
		}
	}
	return nil
}

func set(fm *Frame, r diag.Ranger, temp bool, variable vars.Var, value interface{}) Exception {
	if temp {
		saved := variable.Get()
		err := variable.Set(value)
		if err != nil {
			return fm.errorp(r, err)
		}
		fm.addDefer(func(fm *Frame) Exception {
			err := variable.Set(saved)
			if err != nil {
				return fm.errorpf(r, "restore variable: %w", err)
			}
			return nil
		})
		return nil
	}
	err := variable.Set(value)
	if err != nil {
		return fm.errorp(r, err)
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

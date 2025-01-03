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

func (cp *compiler) compileCompoundLValues(ns []*parse.Compound, f lvalueFlag) lvaluesGroup {
	g := lvaluesGroup{nil, -1}
	for _, n := range ns {
		if len(n.Indexings) != 1 {
			cp.errorpf(n, "lvalue may not be composite expressions")
			break
		}
		more := cp.compileIndexingLValue(n.Indexings[0], f)
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

var dummyLValuesGroup = lvaluesGroup{[]lvalue{{}}, -1}

func (cp *compiler) compileIndexingLValue(n *parse.Indexing, f lvalueFlag) lvaluesGroup {
	if !parse.ValidLHSVariable(n.Head, true) {
		cp.errorpf(n.Head, "lvalue must be valid literal variable names")
		return dummyLValuesGroup
	}
	varUse := n.Head.Value
	sigil, qname := SplitSigil(varUse)
	if qname == "" {
		cp.errorpfPartial(n, "variable name must not be empty")
		return dummyLValuesGroup
	}

	var ref *varRef
	if f&setLValue != 0 {
		ref = resolveVarRef(cp, qname, n)
		if ref != nil && len(ref.subNames) == 0 && ref.info.readOnly {
			cp.errorpf(n, "variable $%s is read-only", parse.Quote(qname))
			return dummyLValuesGroup
		}
	}
	if ref == nil {
		if f&newLValue == 0 {
			cp.autofixUnresolvedVar(qname)
			cp.errorpfPartial(n, "cannot find variable $%s", parse.Quote(qname))
			return dummyLValuesGroup
		}
		if len(n.Indices) > 0 {
			cp.errorpf(n, "new variable $%s must not have indices", parse.Quote(qname))
			return dummyLValuesGroup
		}
		segs := SplitQNameSegs(qname)
		if len(segs) == 1 {
			// Unqualified name - implicit local
			name := segs[0]
			ref = &varRef{localScope,
				staticVarInfo{name, false, false}, cp.thisScope().add(name), nil}
		} else {
			cp.errorpf(n, "cannot create variable $%s; "+
				"new variables can only be created in the current scope",
				parse.Quote(qname))
			return dummyLValuesGroup
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
	var rc restoreCollector
	if op.temp {
		rc = fm.addDefer
	}
	return doAssign(fm, op, op.lhs, op.rhs, rc)
}

func doAssign(fm *Frame, r diag.Ranger, lhs lvaluesGroup, rhs valuesOp, rc restoreCollector) Exception {
	// Evaluate LHS.
	variables := make([]vars.Var, len(lhs.lvalues))
	for i, lvalue := range lhs.lvalues {
		variable, err := derefLValue(fm, lvalue)
		if err != nil {
			return fm.errorp(lhs.lvalues[i], err)
		}
		variables[i] = variable
	}

	// Evaluate RHS.
	values, exc := rhs.exec(fm)
	if exc != nil {
		return exc
	}

	// Now perform assignment.
	if rest := lhs.rest; rest == -1 {
		if len(variables) != len(values) {
			return fm.errorp(r, errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: len(variables), ValidHigh: len(variables), Actual: len(values)})
		}
		for i, variable := range variables {
			exc := set(fm, lhs.lvalues[i], variable, values[i], rc)
			if exc != nil {
				return exc
			}
		}
	} else {
		if len(values) < len(variables)-1 {
			return fm.errorp(r, errs.ArityMismatch{What: "assignment right-hand-side",
				ValidLow: len(variables) - 1, ValidHigh: -1, Actual: len(values)})
		}
		for i := 0; i < rest; i++ {
			exc := set(fm, lhs.lvalues[i], variables[i], values[i], rc)
			if exc != nil {
				return exc
			}
		}
		restOff := len(values) - len(variables)
		exc := set(fm, lhs.lvalues[rest],
			variables[rest], vals.MakeList(values[rest:rest+restOff+1]...), rc)
		if exc != nil {
			return exc
		}
		for i := rest + 1; i < len(variables); i++ {
			exc := set(fm, lhs.lvalues[i], variables[i], values[i+restOff], rc)
			if exc != nil {
				return exc
			}
		}
	}
	return nil
}

type restoreCollector func(func(*Frame) Exception)

// Sets the variable to the value.
//
// If rc is non-empty, calls it with a function that restores the original value
// after setting the variable.
func set(fm *Frame, r diag.Ranger, variable vars.Var, value any, rc restoreCollector) Exception {
	var restore func(*Frame) Exception
	if rc != nil {
		restore = save(r, variable)
	}
	err := variable.Set(value)
	if err != nil {
		return fm.errorp(r, err)
	}
	if rc != nil {
		rc(restore)
	}
	return nil
}

// Returns a function that restores a variable to its current value.
func save(r diag.Ranger, variable vars.Var) func(*Frame) Exception {
	if head := vars.HeadOfElement(variable); head != nil {
		// Needed for temporary assignments to elements (https://b.elv.sh/1515).
		variable = head
	}
	// Handle "unsettable" variables (currently just environment variables)
	// correctly.
	if unsettable, ok := variable.(vars.UnsettableVar); ok && !unsettable.IsSet() {
		return func(fm *Frame) Exception {
			if err := unsettable.Unset(); err != nil {
				return fm.errorpf(r, "unset variable: %w", err)
			}
			return nil
		}
	}
	saved := variable.Get()
	return func(fm *Frame) Exception {
		err := variable.Set(saved)
		if err != nil {
			return fm.errorpf(r, "restore variable: %w", err)
		}
		return nil
	}
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
		return nil, NoSuchVariable(fm.src.Code[lv.From:lv.To])
	}
	if len(lv.indexOps) == 0 {
		return variable, nil
	}
	indices := make([]any, len(lv.indexOps))
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

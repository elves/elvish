package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
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
	qname    string
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
		// Braced list of lvalues may not have indicies.
		if len(n.Indicies) > 0 {
			cp.errorpf(n, "braced list may not have indicies when used as lvalue")
		}
		return cp.parseCompoundLValues(n.Head.Braced)
	}
	// A basic lvalue.
	ref := cp.literal(n.Head, "lvalue only supports literal variable names")
	sigil, qname := SplitVariableRef(ref)
	ends := make([]int, len(n.Indicies)+1)
	ends[0] = n.Head.Range().To
	for i, idx := range n.Indicies {
		ends[i+1] = idx.Range().To
	}
	lv := lvalue{n.Range(), qname, cp.arrayOps(n.Indicies), ends}
	restIndex := -1
	if sigil == "@" {
		restIndex = 0
	}
	// TODO: Deal with other sigils when they exist.
	return lvaluesGroup{[]lvalue{lv}, restIndex}
}

func (cp *compiler) registerLValues(lhs lvaluesGroup) {
	for _, lv := range lhs.lvalues {
		if len(lv.indexOps) == 0 {
			cp.registerVariableSet(lv.qname)
		} else {
			ok := cp.registerVariableGet(lv.qname, lv)
			if !ok {
				cp.errorpf(lv, "variable $%s not found", lv.qname)
			}
		}
	}
}

type assignOp struct {
	lhs lvaluesGroup
	rhs valuesOp
}

func (op *assignOp) invoke(fm *Frame) error {
	variables := make([]vars.Var, len(op.lhs.lvalues))
	for i, lvalue := range op.lhs.lvalues {
		variable, err := getVar(fm, lvalue)
		if err != nil {
			return err
		}
		variables[i] = variable
	}

	values, err := op.rhs.exec(fm)
	if err != nil {
		return err
	}

	if op.lhs.rest == -1 {
		if len(variables) != len(values) {
			return errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: len(variables), ValidHigh: len(variables), Actual: len(values)}
		}
		for i, variable := range variables {
			err := variable.Set(values[i])
			if err != nil {
				return err
			}
		}
	} else {
		if len(values) < len(variables)-1 {
			return errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: len(variables) - 1, ValidHigh: -1, Actual: len(values)}
		}
		rest := op.lhs.rest
		for i := 0; i < rest; i++ {
			err := variables[i].Set(values[i])
			if err != nil {
				return err
			}
		}
		restOff := len(values) - len(variables)
		err := variables[rest].Set(vals.MakeList(values[rest : rest+restOff+1]...))
		if err != nil {
			return err
		}
		for i := rest + 1; i < len(variables); i++ {
			err := variables[i].Set(values[i+restOff])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getVar(fm *Frame, lv lvalue) (vars.Var, error) {
	variable := fm.ResolveVar(lv.qname)
	if variable == nil {
		ns, _ := SplitQNameNs(lv.qname)
		if ns == "" || ns == ":" || ns == "local:" {
			// This should have been created as part of pipelineOp.
			return nil, errors.New("compiler bug: new local variable not created in pipeline")
		}
		return nil, fmt.Errorf("new variables can only be created in local scope")
	}
	if len(lv.indexOps) == 0 {
		return variable, nil
	}
	indicies := make([]interface{}, len(lv.indexOps))
	for i, op := range lv.indexOps {
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
			return nil, fm.errorp(lv, err)
		}
		return nil, fm.errorp(diag.Ranging{From: lv.From, To: lv.ends[level]}, err)
	}
	return elemVar, nil
}

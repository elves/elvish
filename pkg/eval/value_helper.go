package eval

import (
	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vars"
)

func evalForValue(fm *Frame, op valuesOp, what string) (interface{}, error) {
	values, err := op.exec(fm)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fm.errorp(op, errs.ArityMismatch{
			What: what, ValidLow: 1, ValidHigh: 1, Actual: len(values)})
	}
	return values[0], nil
}

func evalForVar(fm *Frame, op lvaluesOp, what string) (vars.Var, error) {
	lvalues, err := op.exec(fm)
	if err != nil {
		return nil, err
	}
	if len(lvalues) != 1 {
		return nil, fm.errorp(op, errs.ArityMismatch{
			What: what, ValidLow: 1, ValidHigh: 1, Actual: len(lvalues)})
	}
	return lvalues[0], nil
}

package eval

import (
	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vars"
)

// Op represents an operation on a Frame. It is the result of compiling a piece
// of source.
type Op struct {
	Inner effectOp
	Src   *Source
}

// An operation on a Frame that produces a side effect.
type effectOp struct {
	body effectOpBody
	diag.Ranging
}

func makeEffectOp(r diag.Ranger, body effectOpBody) effectOp {
	return effectOp{body, r.Range()}
}

// The body of an effectOp.
type effectOpBody interface {
	invoke(*Frame) error
}

// Executes an effectOp for side effects.
func (op effectOp) exec(fm *Frame) error {
	fm.begin, fm.end = op.From, op.To
	return op.body.invoke(fm)
}

// An operation on an Frame that produce Value's.
type valuesOp struct {
	body valuesOpBody
	diag.Ranging
}

func makeValuesOp(r diag.Ranger, body valuesOpBody) valuesOp {
	return valuesOp{body, r.Range()}
}

// The body of ValuesOp.
type valuesOpBody interface {
	invoke(*Frame) ([]interface{}, error)
}

// Executes a ValuesOp and produces values.
func (op valuesOp) exec(fm *Frame) ([]interface{}, error) {
	fm.begin, fm.end = op.From, op.To
	return op.body.invoke(fm)
}

// An operation on a Frame that produce Variable's.
type lvaluesOp struct {
	body lvaluesOpBody
	diag.Ranging
}

// The body of an LValuesOp.
type lvaluesOpBody interface {
	invoke(*Frame) ([]vars.Var, error)
}

// Executes an LValuesOp and produces lvalues.
func (op lvaluesOp) exec(fm *Frame) ([]vars.Var, error) {
	// Empty value is considered to generate no lvalues.
	if op.body == nil {
		return []vars.Var{}, nil
	}
	fm.begin, fm.end = op.From, op.To
	return op.body.invoke(fm)
}

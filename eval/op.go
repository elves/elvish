package eval

import "github.com/elves/elvish/eval/vars"

// Op represents an operation on a Frame. It is the result of compiling a piece
// of source.
type Op struct {
	Inner effectOp
	Src   *Source
}

// An operation on a Frame that produces a side effect.
type effectOp struct {
	body       opBody
	begin, end int
}

// The body of an Op.
type opBody interface {
	invoke(*Frame) error
}

// Executes an effectOp for side effects.
func (op effectOp) exec(fm *Frame) error {
	fm.begin, fm.end = op.begin, op.end
	return op.body.invoke(fm)
}

// An operation on an Frame that produce Value's.
type valuesOp struct {
	body       valuesOpBody
	begin, end int
}

// The body of ValuesOp.
type valuesOpBody interface {
	invoke(*Frame) ([]interface{}, error)
}

// Executes a ValuesOp and produces values.
func (op valuesOp) exec(fm *Frame) ([]interface{}, error) {
	fm.begin, fm.end = op.begin, op.end
	return op.body.invoke(fm)
}

// An operation on a Frame that produce Variable's.
type lvaluesOp struct {
	body       lvaluesOpBody
	begin, end int
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
	fm.begin, fm.end = op.begin, op.end
	return op.body.invoke(fm)
}

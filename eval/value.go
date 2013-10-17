package eval

import (
	"fmt"
	"bytes"
)

type Value interface {
	meisvalue()
	String() string
	Caret(v Value) Value
}

// TODO Only str part is used.
type Scalar struct {
	num float64
	str string
}
func (s *Scalar) meisvalue() {}

func NewScalar(s string) *Scalar {
	return &Scalar{str: s}
}

func (s *Scalar) String() string {
	return s.str
}

func (s *Scalar) Caret(v Value) Value {
	return NewScalar(s.str + v.String())
}

type Table struct {
	list []Value
	dict map[Value]Value
}
func (t *Table) meisvalue() {}

func NewTable() *Table {
	return &Table{dict: make(map[Value]Value)}
}

func (t *Table) String() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	sep := ""
	for _, v := range t.list {
		fmt.Fprint(buf, sep, v.String())
		sep = " "
	}
	for k, v := range t.dict {
		fmt.Fprint(buf, sep, k.String(), " = ", v.String())
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (t *Table) Caret(v Value) Value {
	// TODO Implement indexing
	return NewScalar(t.String() + v.String())
}

func (t *Table) append(vs... Value) {
	t.list = append(t.list, vs...)
}

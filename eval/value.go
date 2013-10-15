package eval

import (
	"fmt"
	"bytes"
)

type Value interface {
	meisvalue()
	String() string
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

// TODO Not yet used.
type Table struct {
	list []Value
	dict map[Value]Value
}
func (t *Table) meisvalue() {}

func NewTable() *Table {
	return &Table{make([]Value, 0), make(map[Value]Value)}
}

func (t *Table) String() string {
	buf := new(bytes.Buffer)
	sep := '['
	for _, v := range t.list {
		fmt.Fprint(buf, sep, v.String())
		sep = ' '
	}
	for k, v := range t.dict {
		fmt.Fprint(buf, sep, '(', k.String(), ' ', v.String())
		sep = ' '
	}
	buf.WriteRune(']')
	return buf.String()
}

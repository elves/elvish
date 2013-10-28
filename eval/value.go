package eval

import (
	"fmt"
	"bytes"
	"strings"
	"strconv"
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
		fmt.Fprint(buf, sep, "&", k.String(), " ", v.String())
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (t *Table) Caret(v Value) Value {
	switch v := v.(type) {
	case *Scalar:
		return NewScalar(t.String() + v.String())
	case *Table:
		if len(v.list) != 1 || len(v.dict) != 0 {
			// TODO Use Evaluator.errorf
			panic("subscription must be single-element list")
		}
		sub, ok := v.list[0].(*Scalar)
		if !ok {
			// TODO Use Evaluator.errorf
			panic("subscription must be single-element scalar list")
		}
		// Need stricter notion of list indices
		// TODO Handle invalid index
		idx, err := strconv.ParseUint(sub.String(), 10, 0)
		if err == nil {
			return t.list[idx]
		} else {
			return t.dict[sub]
		}
	default:
		// TODO Use Evaluator.errorf
		panic("Table can only be careted with Scalar or Table")
	}
}

func (t *Table) append(vs... Value) {
	t.list = append(t.list, vs...)
}

type Env struct {
	m map[string]string
}
func (e *Env) meisvalue() {}

func NewEnv(s []string) *Env {
	e := &Env{make(map[string]string)}
	for _, s := range s {
		arr := strings.SplitN(s, "=", 2)
		if len(arr) == 2 {
			e.m[arr[0]] = arr[1]
		}
	}
	return e
}

func (e *Env) Export() []string {
	s := make([]string, 0, len(e.m))
	for k, v := range e.m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

func (e *Env) String() string {
	return "$env"
}

func (e *Env) Caret(v Value) Value {
	switch v := v.(type) {
	case *Table:
		if len(v.list) != 1 || len(v.dict) != 0 {
			// TODO Use Evaluator.errorf
			panic("subscription must be single-element list")
		}
		sub, ok := v.list[0].(*Scalar)
		if !ok {
			// TODO Use Evaluator.errorf
			panic("subscription must be single-element scalar list")
		}
		// TODO Handle invalid index
		return NewScalar(e.m[sub.String()])
	default:
		panic("Env can only be careted with Table")
	}
}

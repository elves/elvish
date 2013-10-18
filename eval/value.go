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
		fmt.Fprint(buf, sep, k.String(), " = ", v.String())
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
		idx, err := strconv.ParseUint(sub.String(), 10, 0)
		if err == nil {
			return t.list[idx]
		} else {
			return t.dict[sub]
		}
	default:
		panic("unreachable")
	}
}

func (t *Table) append(vs... Value) {
	t.list = append(t.list, vs...)
}

func envAsSlice(env map[string]string) (s []string) {
	s = make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func envAsMap(env []string) (m map[string]string) {
	m = make(map[string]string)
	for _, e := range env {
		arr := strings.SplitN(e, "=", 2)
		if len(arr) == 2 {
			m[arr[0]] = arr[1]
		}
	}
	return
}

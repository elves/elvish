package eval

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type Type interface {
	Default() Value
}

type AnyType struct {
}

func (at AnyType) Default() Value {
	return NewString("")
}

type StringType struct {
}

func (st StringType) Default() Value {
	return NewString("")
}

type TableType struct {
}

func (tt TableType) Default() Value {
	return NewTable()
}

type EnvType struct {
}

func (et EnvType) Default() Value {
	return NewEnv()
}

type ClosureType struct {
	Bounds [2]StreamType
}

func (st ClosureType) Default() Value {
	return NewClosure([]string{}, nil, map[string]*Value{}, st.Bounds)
}

var typenames = map[string]Type{
	"string":  StringType{},
	"table":   TableType{},
	"env":     EnvType{},
	"closure": ClosureType{[2]StreamType{}},
}

// Value is the runtime representation of an elvish value.
type Value interface {
	Type() Type
	Repr() string
	String() string
}

func valuePtr(v Value) *Value {
	return &v
}

// String is a string.
type String string

func (s *String) Type() Type {
	return StringType{}
}

func NewString(s string) *String {
	ss := String(s)
	return &ss
}

func quote(s string) string {
	if len(s) == 0 {
		return "``"
	}

	printable := true
	for _, r := range s {
		if !unicode.IsPrint(r) {
			printable = false
			break
		}
	}
	if printable {
		r0, w0 := utf8.DecodeRuneInString(s)
		if parse.StartsBare(r0) {
			barewordPossible := true
			for _, r := range s[w0:] {
				if parse.TerminatesBare(r) {
					barewordPossible = false
					break
				}
			}
			if barewordPossible {
				return s
			}
		}

		// Quote with backquote
		buf := new(bytes.Buffer)
		buf.WriteRune('`')
		for _, r := range s {
			buf.WriteRune(r)
			if r == '`' {
				buf.WriteRune('`')
			}
		}
		buf.WriteRune('`')
		return buf.String()
	}
	// Quote with double quote
	return strconv.Quote(s)
}

func (s *String) Repr() string {
	return quote(string(*s))
}

func (s *String) String() string {
	return string(*s)
}

// Table is a list-dict hybrid.
//
// TODO(xiaq): The dict part use string keys. It should use Value keys instead.
type Table struct {
	List []Value
	Dict map[string]Value
}

func (t *Table) Type() Type {
	return TableType{}
}

func NewTable() *Table {
	return &Table{Dict: make(map[string]Value)}
}

func (t *Table) Repr() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	sep := ""
	for _, v := range t.List {
		fmt.Fprint(buf, sep, v.Repr())
		sep = " "
	}
	for k, v := range t.Dict {
		fmt.Fprint(buf, sep, "&", quote(k), " ", v.Repr())
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (t *Table) String() string {
	return t.Repr()
}

func (t *Table) append(vs ...Value) {
	t.List = append(t.List, vs...)
}

// Env provides access to environment variables.
type Env struct {
	m map[string]string
}

func (e *Env) Type() Type {
	return EnvType{}
}

func NewEnv() *Env {
	return &Env{}
}

func (e *Env) fill() {
	if e.m != nil {
		return
	}
	e.m = make(map[string]string)
	for _, s := range os.Environ() {
		arr := strings.SplitN(s, "=", 2)
		if len(arr) == 2 {
			e.m[arr[0]] = arr[1]
		}
	}
}

func (e *Env) Export() []string {
	e.fill()
	s := make([]string, 0, len(e.m))
	for k, v := range e.m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

func (e *Env) Repr() string {
	e.fill()
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	sep := ""
	for k, v := range e.m {
		fmt.Fprint(buf, sep, "&", quote(k), " ", quote(v))
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (e *Env) String() string {
	e.fill()
	return e.Repr()
}

// Closure is a closure.
type Closure struct {
	ArgNames []string
	Op       Op
	Enclosed map[string]*Value
	Bounds   [2]StreamType
}

func (c *Closure) Type() Type {
	return ClosureType{c.Bounds}
}

func NewClosure(a []string, op Op, e map[string]*Value, b [2]StreamType) *Closure {
	return &Closure{a, op, e, b}
}

func (c *Closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

func (c *Closure) String() string {
	return c.Repr()
}

func evalSubscript(ev *Evaluator, left, right Value, lp, rp parse.Pos) Value {
	var (
		sub *String
		ok  bool
	)
	if sub, ok = right.(*String); !ok {
		ev.errorf(rp, "right operand of subscript must be of type string")
	}

	switch left.(type) {
	case *Env:
		return NewString(left.(*Env).m[sub.String()])
	case *Table:
		t := left.(*Table)
		// Need stricter notion of list indices
		// TODO Handle invalid index
		idx, err := strconv.ParseUint(sub.String(), 10, 0)
		if err == nil {
			if idx < uint64(len(t.List)) {
				return t.List[idx]
			}
			ev.errorf(rp, "index out of range")
		}
		if v, ok := t.Dict[sub.String()]; ok {
			return v
		}
		ev.errorf(rp, "nonexistent key %q", sub)
		return nil
	case *String:
		invalidIndex := "invalid index, must be integer or integer:integer"

		ss := strings.Split(sub.String(), ":")
		if len(ss) > 2 {
			ev.errorf(rp, invalidIndex)
		}
		idx := make([]int, len(ss))
		for i, s := range ss {
			n, err := strconv.ParseInt(s, 10, 0)
			if err != nil {
				ev.errorf(rp, invalidIndex)
			}
			idx[i] = int(n)
		}

		var s string
		var e error
		if len(idx) == 1 {
			var r rune
			r, e = util.NthRune(left.String(), idx[0])
			s = string(r)
		} else {
			s, e = util.SubstringByRune(left.String(), idx[0], idx[1])
		}
		if e != nil {
			ev.errorf(rp, "%v", e)
		}
		return NewString(s)
	default:
		ev.errorf(lp, "left operand of subscript must be of type string, env, table or any")
		return nil
	}
}

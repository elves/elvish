package eval

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/xiaq/elvish/parse"
)

type Type interface {
	Default() Value
	Caret(t Type) Type
}

type AnyType struct {
}

func (at AnyType) Default() Value {
	return NewString("")
}

func (at AnyType) Caret(t Type) Type {
	return AnyType{}
}

type StringType struct {
}

func (st StringType) Default() Value {
	return NewString("")
}

func (st StringType) Caret(t Type) Type {
	return StringType{}
}

type TableType struct {
}

func (tt TableType) Default() Value {
	return NewTable()
}

func (tt TableType) Caret(t Type) Type {
	// TODO Should check t
	return AnyType{}
}

type EnvType struct {
}

func (et EnvType) Default() Value {
	return NewEnv()
}

func (et EnvType) Caret(t Type) Type {
	// TODO Should check t
	return StringType{}
}

type ClosureType struct {
	Bounds [2]StreamType
}

func (st ClosureType) Default() Value {
	return NewClosure([]string{}, nil, map[string]*Value{}, st.Bounds)
}

func (ct ClosureType) Caret(t Type) Type {
	// TODO Should always be a Compiler error
	return AnyType{}
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
	Caret(ev *Evaluator, v Value) Value
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

func (s *String) Caret(ev *Evaluator, v Value) Value {
	return NewString(string(*s) + v.String())
}

// Table is a list-dict hybrid.
type Table struct {
	List []Value
	Dict map[Value]Value
}

func (t *Table) Type() Type {
	return TableType{}
}

func NewTable() *Table {
	return &Table{Dict: make(map[Value]Value)}
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
		fmt.Fprint(buf, sep, "&", k.Repr(), " ", v.Repr())
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (t *Table) String() string {
	return t.Repr()
}

func (t *Table) Caret(ev *Evaluator, v Value) Value {
	switch v := v.(type) {
	case *String:
		return NewString(t.String() + v.String())
	case *Table:
		if len(v.List) != 1 || len(v.Dict) != 0 {
			ev.errorf("subscription must be single-element list")
		}
		sub, ok := v.List[0].(*String)
		if !ok {
			ev.errorf("subscription must be single-element string list")
		}
		// Need stricter notion of list indices
		// TODO Handle invalid index
		idx, err := strconv.ParseUint(sub.String(), 10, 0)
		if err == nil {
			return t.List[idx]
		}
		return t.Dict[sub]
	default:
		ev.errorf("Table can only be careted with String or Table")
		return nil
	}
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

func (e *Env) Caret(ev *Evaluator, v Value) Value {
	e.fill()
	switch v := v.(type) {
	case *Table:
		if len(v.List) != 1 || len(v.Dict) != 0 {
			ev.errorf("subscription must be single-element list")
		}
		sub, ok := v.List[0].(*String)
		if !ok {
			ev.errorf("subscription must be single-element string list")
		}
		// TODO Handle invalid index
		return NewString(e.m[sub.String()])
	default:
		ev.errorf("Env can only be careted with Table")
		return nil
	}
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

func (c *Closure) Caret(ev *Evaluator, v Value) Value {
	ev.errorf("Closure doesn't support careting")
	return nil
}

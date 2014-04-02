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
	return NewClosure(
		[]string{}, &parse.ChunkNode{0, []*parse.PipelineNode{}},
		map[string]*Value{}, st.Bounds)
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
	Repr(ev *Evaluator) string
	String(ev *Evaluator) string
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

func (s *String) Repr(ev *Evaluator) string {
	return quote(string(*s))
}

func (s *String) String(ev *Evaluator) string {
	return string(*s)
}

func (s *String) Caret(ev *Evaluator, v Value) Value {
	return NewString(string(*s) + v.String(ev))
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

func (t *Table) Repr(ev *Evaluator) string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	sep := ""
	for _, v := range t.List {
		fmt.Fprint(buf, sep, v.Repr(ev))
		sep = " "
	}
	for k, v := range t.Dict {
		fmt.Fprint(buf, sep, "&", k.Repr(ev), " ", v.Repr(ev))
		sep = " "
	}
	buf.WriteRune(']')
	return buf.String()
}

func (t *Table) String(ev *Evaluator) string {
	return t.Repr(ev)
}

func (t *Table) Caret(ev *Evaluator, v Value) Value {
	switch v := v.(type) {
	case *String:
		return NewString(t.String(ev) + v.String(ev))
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
		idx, err := strconv.ParseUint(sub.String(ev), 10, 0)
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

func (e *Env) Repr(ev *Evaluator) string {
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

func (e *Env) String(ev *Evaluator) string {
	e.fill()
	return e.Repr(ev)
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
		return NewString(e.m[sub.String(ev)])
	default:
		ev.errorf("Env can only be careted with Table")
		return nil
	}
}

// Closure is a closure.
type Closure struct {
	ArgNames []string
	Chunk    *parse.ChunkNode
	Enclosed map[string]*Value
	Bounds   [2]StreamType
}

func (c *Closure) Type() Type {
	return ClosureType{c.Bounds}
}

func NewClosure(a []string, ch *parse.ChunkNode, e map[string]*Value, b [2]StreamType) *Closure {
	return &Closure{a, ch, e, b}
}

func (c *Closure) Repr(ev *Evaluator) string {
	return fmt.Sprintf("<Closure%v>", c)
}

func (c *Closure) String(ev *Evaluator) string {
	return c.Repr(ev)
}

func (c *Closure) Caret(ev *Evaluator, v Value) Value {
	ev.errorf("Closure doesn't support careting")
	return nil
}

package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/xiaq/elvish/parse"
)

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

// Value is the runtime representation of an elvish value.
type Value interface {
	isValue()
	Repr(ev *Evaluator) string
	String(ev *Evaluator) string
	Caret(ev *Evaluator, v Value) Value
}

// TODO Only str part is used.

// String is a string.
type String string

func (s *String) isValue() {}

func NewString(s string) *String {
	ss := String(s)
	return &ss
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

func (t *Table) isValue() {}

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

// Env is a special-cased Table. The only instance of it is the global $env.
type Env struct {
	m map[string]string
}

func (e *Env) isValue() {}

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

func (e *Env) Repr(ev *Evaluator) string {
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
	return e.Repr(ev)
}

func (e *Env) Caret(ev *Evaluator, v Value) Value {
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
	Chunk    *parse.ListNode
}

func (c *Closure) isValue() {}

func NewClosure(argNames []string, ch *parse.ListNode) *Closure {
	return &Closure{argNames, ch}
}

func (c *Closure) Repr(ev *Evaluator) string {
	return fmt.Sprintf("<closure %p>", c.Chunk)
}

func (c *Closure) String(ev *Evaluator) string {
	return c.Repr(ev)
}

func (c *Closure) Caret(ev *Evaluator, v Value) Value {
	ev.errorf("Closure doesn't support careting")
	return nil
}

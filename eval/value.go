package eval

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Value is the runtime representation of an elvish value.
type Value interface {
	Type() Type
	Repr() string
}

type Booler interface {
	Bool() bool
}

type Stringer interface {
	String() string
}

// String is a string.
type String string

func (s String) Type() Type {
	return StringType{}
}

func NewString(s string) String {
	return String(s)
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

func (s String) Repr() string {
	return quote(string(s))
}

func (s String) String() string {
	return string(s)
}

type Bool bool

func (b Bool) Type() Type {
	return BoolType{}
}

func (b Bool) Repr() string {
	if b {
		return "$true"
	} else {
		return "$false"
	}
}

func (b Bool) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func (b Bool) Bool() bool {
	return bool(b)
}

type Exitus struct {
	Success bool
	Failure string
}

var success = Exitus{true, ""}

func newFailure(s string) Exitus {
	return Exitus{false, s}
}

func (e Exitus) Type() Type {
	return ExitusType{}
}

func (e Exitus) Repr() string {
	if e.Success {
		return "$success"
	} else {
		return "(failure " + quote(e.Failure) + ")"
	}
}

func (e Exitus) String() string {
	if e.Success {
		return "success"
	} else {
		return "failure: " + e.Failure
	}
}

func (e Exitus) Bool() bool {
	return e.Success
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

func (t *Table) append(vs ...Value) {
	t.List = append(t.List, vs...)
}

type Callable interface {
	Value
	// Exec executes a callable asynchronously on an Evaluator. It assumes that
	// it is the last callable to be executed on that Evaluator and thus
	// responsible for cleaning up the ports.
	Exec(ev *Evaluator, args []Value) <-chan *StateUpdate
}

// Closure is a closure.
type Closure struct {
	ArgNames []string
	Op       Op
	Captured map[string]Variable
}

func (c *Closure) Type() Type {
	return CallableType{}
}

func NewClosure(a []string, op Op, e map[string]Variable) *Closure {
	return &Closure{a, op, e}
}

func (c *Closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

type BuiltinFn struct {
	Name string
	Impl func(*Evaluator, []Value) Exitus
}

func (b *BuiltinFn) Type() Type {
	return CallableType{}
}

func (b *BuiltinFn) Repr() string {
	return "$builtin:fn-" + b.Name
}

type External struct {
	Name string
}

func (e External) Type() Type {
	return CallableType{}
}

func (e External) Repr() string {
	return "<external " + e.Name + " >"
}

func evalSubscript(ev *Evaluator, left, right Value, lp, rp parse.Pos) Value {
	var (
		sub String
		ok  bool
	)
	if sub, ok = right.(String); !ok {
		ev.errorf(rp, "right operand of subscript must be of type string")
	}

	switch left.(type) {
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
	case String:
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
			r, e = util.NthRune(toString(left), idx[0])
			s = string(r)
		} else {
			s, e = util.SubstringByRune(toString(left), idx[0], idx[1])
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

// fromJSONInterface converts a interface{} that results from json.Unmarshal to
// a Value.
func fromJSONInterface(v interface{}) Value {
	if v == nil {
		// TODO Use a more appropriate type
		return String("")
	}
	switch v.(type) {
	case bool:
		return Bool(v.(bool))
	case float64, string:
		// TODO Use a numeric type for float64
		return String(fmt.Sprint(v))
	case []interface{}:
		a := v.([]interface{})
		t := &Table{make([]Value, len(a)), make(map[string]Value)}
		for i, v := range a {
			t.List[i] = fromJSONInterface(v)
		}
		return t
	case map[string]interface{}:
		m := v.(map[string]interface{})
		t := NewTable()
		for k, v := range m {
			t.Dict[k] = fromJSONInterface(v)
		}
		return t
	default:
		// TODO Find a better way to report error
		return newFailure(fmt.Sprintf("unexpected json type: %T", v))
	}
}

func valueEq(a, b Value) bool {
	// XXX(xiaq): This is cheating. May no longer be true after values get more
	// complex.
	return reflect.DeepEqual(a, b)
}

// toString converts a Value to String. When the Value type implements
// String(), it is used. Otherwise Repr() is used.
func toString(v Value) string {
	if s, ok := v.(Stringer); ok {
		return s.String()
	}
	return v.Repr()
}

// toBool converts a Value to bool. When the Value type implements Bool(), it
// is used. Otherwise it is considered true.
func toBool(v Value) bool {
	if b, ok := v.(Booler); ok {
		return b.Bool()
	}
	return true
}

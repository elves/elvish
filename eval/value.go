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

type booler interface {
	Bool() bool
}

type stringer interface {
	String() string
}

type str string

func (s str) Type() Type {
	return stringType{}
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

func (s str) Repr() string {
	return quote(string(s))
}

func (s str) String() string {
	return string(s)
}

type boolean bool

func (b boolean) Type() Type {
	return boolType{}
}

func (b boolean) Repr() string {
	if b {
		return "$true"
	}
	return "$false"
}

func (b boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b boolean) Bool() bool {
	return bool(b)
}

type exitus struct {
	Success bool
	Failure string
}

var success = exitus{true, ""}

func newFailure(s string) exitus {
	return exitus{false, s}
}

func (e exitus) Type() Type {
	return exitusType{}
}

func (e exitus) Repr() string {
	if e.Success {
		return "$success"
	}
	return "(failure " + quote(e.Failure) + ")"
}

func (e exitus) String() string {
	if e.Success {
		return "success"
	}
	return "failure: " + e.Failure
}

func (e exitus) Bool() bool {
	return e.Success
}

// table is a list-dict hybrid.
//
// TODO(xiaq): The dict part use string keys. It should use Value keys instead.
type table struct {
	List []Value
	Dict map[string]Value
}

func (t *table) Type() Type {
	return tableType{}
}

func newTable() *table {
	return &table{Dict: make(map[string]Value)}
}

func (t *table) Repr() string {
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

func (t *table) append(vs ...Value) {
	t.List = append(t.List, vs...)
}

// Callable represents Value's that may be executed.
type callable interface {
	Value
	// Exec executes a callable asynchronously on an Evaluator. It assumes that
	// it is the last callable to be executed on that Evaluator and thus
	// responsible for cleaning up the ports.
	Exec(ev *Evaluator, args []Value) <-chan *stateUpdate
}

// closure is a closure.
type closure struct {
	ArgNames []string
	Op       op
	Captured map[string]Variable
}

func (c *closure) Type() Type {
	return callableType{}
}

func newClosure(a []string, op op, e map[string]Variable) *closure {
	return &closure{a, op, e}
}

func (c *closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

type builtinFn struct {
	Name string
	Impl func(*Evaluator, []Value) exitus
}

func (b *builtinFn) Type() Type {
	return callableType{}
}

func (b *builtinFn) Repr() string {
	return "$builtin:fn-" + b.Name
}

type externalCmd struct {
	Name string
}

func (e externalCmd) Type() Type {
	return callableType{}
}

func (e externalCmd) Repr() string {
	return "<external " + e.Name + " >"
}

func evalSubscript(ev *Evaluator, left, right Value, lp, rp parse.Pos) Value {
	var (
		sub str
		ok  bool
	)
	if sub, ok = right.(str); !ok {
		ev.errorf(rp, "right operand of subscript must be of type string")
	}

	switch left.(type) {
	case *table:
		t := left.(*table)
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
	case str:
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
		return str(s)
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
		return str("")
	}
	switch v.(type) {
	case bool:
		return boolean(v.(bool))
	case float64, string:
		// TODO Use a numeric type for float64
		return str(fmt.Sprint(v))
	case []interface{}:
		a := v.([]interface{})
		t := &table{make([]Value, len(a)), make(map[string]Value)}
		for i, v := range a {
			t.List[i] = fromJSONInterface(v)
		}
		return t
	case map[string]interface{}:
		m := v.(map[string]interface{})
		t := newTable()
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
	if s, ok := v.(stringer); ok {
		return s.String()
	}
	return v.Repr()
}

// toBool converts a Value to bool. When the Value type implements Bool(), it
// is used. Otherwise it is considered true.
func toBool(v Value) bool {
	if b, ok := v.(booler); ok {
		return b.Bool()
	}
	return true
}

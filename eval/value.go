package eval

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/strutil"
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

type traceback struct {
	// TODO(xiaq): Add context information
	causes []exitus
}

type exitusSort byte

const (
	Ok exitusSort = iota
	Failure
	Traceback

	// Control flow sorts
	Return
	Break
	Continue
	FlowSortLower = Return
)

var flowExitusNames = map[exitusSort]string{
	Return: "return", Break: "break", Continue: "continue",
}

type exitus struct {
	Sort      exitusSort
	Failure   string
	Traceback *traceback
}

var ok = exitus{Ok, "", nil}

func newTraceback(es []exitus) exitus {
	return exitus{Traceback, "", &traceback{es}}
}

func newFailure(s string) exitus {
	return exitus{Failure, s, nil}
}

func newFlowExitus(s exitusSort) exitus {
	return exitus{s, "", nil}
}

func (e exitus) Type() Type {
	return exitusType{}
}

func (e exitus) Repr() string {
	switch e.Sort {
	case Ok:
		return "$ok"
	case Failure:
		return "(failure " + quote(e.Failure) + ")"
	case Traceback:
		b := new(bytes.Buffer)
		b.WriteString("(traceback")
		for _, c := range e.Traceback.causes {
			b.WriteString(" ")
			b.WriteString(c.Repr())
		}
		b.WriteString(")")
		return b.String()
	default:
		return "?(" + flowExitusNames[e.Sort] + ")"
	}
}

func (e exitus) String() string {
	switch e.Sort {
	case Ok:
		return "ok"
	case Failure:
		return "failure: " + e.Failure
	case Traceback:
		b := new(bytes.Buffer)
		b.WriteString("traceback: (")
		for i, c := range e.Traceback.causes {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(c.String())
		}
		b.WriteString(")")
		return b.String()
	default:
		return flowExitusNames[e.Sort]
	}
}

func (e exitus) Bool() bool {
	return e.Sort == Ok
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
	// Exec executes a callable asynchronously on an Evaler. It assumes that
	// it is the last callable to be executed on that Evaler and thus
	// responsible for cleaning up the ports.
	Exec(ec *evalCtx, args []Value) <-chan *stateUpdate
}

// closure is a closure.
type closure struct {
	ArgNames []string
	Op       valuesOp
	Captured map[string]Variable
}

func (c *closure) Type() Type {
	return callableType{}
}

func newClosure(a []string, op valuesOp, e map[string]Variable) *closure {
	return &closure{a, op, e}
}

func (c *closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

type builtinFn struct {
	Name string
	Impl func(*evalCtx, []Value) exitus
}

func (b *builtinFn) Type() Type {
	return callableType{}
}

func (b *builtinFn) Repr() string {
	return "$builtin:" + fnPrefix + b.Name
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

type rat struct {
	b *big.Rat
}

func newRat() rat {
	return rat{&big.Rat{}}
}

func (r rat) Type() Type {
	return ratType{}
}

func (r rat) Repr() string {
	return "(rat " + r.String() + ")"
}

func (r rat) String() string {
	if r.b.IsInt() {
		return r.b.Num().String()
	}
	return r.b.String()
}

func evalSubscript(ec *evalCtx, left, right Value, lp, rp parse.Pos) Value {
	var sub string
	if s, ok := right.(str); ok {
		sub = string(s)
	} else {
		ec.errorf(rp, "right operand of subscript must be of type string")
	}

	switch left.(type) {
	case *table:
		t := left.(*table)
		// TODO(xiaq): An index is considered a list index if it can be parsed
		// as an unsigned integer; otherwise it is a dict index. This is
		// somewhat subtle.
		idx, err := strconv.ParseUint(sub, 10, 0)
		if err == nil {
			if idx < uint64(len(t.List)) {
				return t.List[idx]
			}
			ec.errorf(rp, "index out of range")
		}
		if v, ok := t.Dict[sub]; ok {
			return v
		}
		ec.errorf(rp, "nonexistent key %q", sub)
		return nil
	case str:
		invalidIndex := "invalid index, must be integer or integer:integer"

		ss := strings.Split(sub, ":")
		if len(ss) > 2 {
			ec.errorf(rp, invalidIndex)
		}
		idx := make([]int, len(ss))
		for i, s := range ss {
			n, err := strconv.ParseInt(s, 10, 0)
			if err != nil {
				ec.errorf(rp, invalidIndex)
			}
			idx[i] = int(n)
		}

		var s string
		var e error
		if len(idx) == 1 {
			var r rune
			r, e = strutil.NthRune(toString(left), idx[0])
			s = string(r)
		} else {
			s, e = strutil.SubstringByRune(toString(left), idx[0], idx[1])
		}
		if e != nil {
			ec.errorf(rp, "%v", e)
		}
		return str(s)
	default:
		ec.errorf(lp, "left operand of subscript must be of type string, env, table or any")
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
	// BUG(xiaq): valueEq uses reflect.DeepEqual to check the equality of two
	// values, may can become wrong when values get more complex.
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

func allTrue(vs []Value) bool {
	for _, v := range vs {
		if !toBool(v) {
			return false
		}
	}
	return true
}

var errOnlyStrOrRat = errors.New("only str or rat may be converted to rat")

// toRat converts a Value to rat. A str can be converted to a rat if it can be
// parsed. A rat is returned as-is. Other types of values cannot be converted.
func toRat(v Value) (rat, error) {
	switch v := v.(type) {
	case rat:
		return v, nil
	case str:
		r := big.Rat{}
		_, err := fmt.Sscanln(string(v), &r)
		if err != nil {
			return rat{}, fmt.Errorf("%s cannot be parsed as rat", v.Repr())
		}
		return rat{&r}, nil
	default:
		return rat{}, errOnlyStrOrRat
	}
}

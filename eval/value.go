package eval

//go:generate stringer -type=Type

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

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

type indexer interface {
	Index(idx string) (Value, error)
}

type Type int

const (
	invalidType Type = iota
	stringType
	exitusType
	boolType
	listType
	mapType
	callableType
	ratType
)

type str string

func (s str) Type() Type {
	return stringType
}

func (s str) Repr() string {
	return parse.Quote(string(s))
}

func (s str) String() string {
	return string(s)
}

var (
	needIntIndex = errors.New("need integer index")
)

func (s str) Index(idx string) (Value, error) {
	i, err := strconv.Atoi(idx)
	if err != nil {
		return nil, err
	}

	r, err := strutil.NthRune(string(s), i)
	if err != nil {
		return nil, err
	}
	return str(string(r)), nil
}

type boolean bool

func (b boolean) Type() Type {
	return boolType
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
	causes []Exitus
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

type Exitus struct {
	Sort      exitusSort
	Failure   string
	Traceback *traceback
}

var (
	OK             = Exitus{Ok, "", nil}
	GenericFailure = Exitus{Failure, "generic failure", nil}
)

func newTraceback(es ...Exitus) Exitus {
	return Exitus{Traceback, "", &traceback{es}}
}

func NewFailure(s string) Exitus {
	return Exitus{Failure, s, nil}
}

func newFlowExitus(s exitusSort) Exitus {
	return Exitus{s, "", nil}
}

func (e Exitus) Type() Type {
	return exitusType
}

func (e Exitus) Repr() string {
	switch e.Sort {
	case Ok:
		return "$ok"
	case Failure:
		return "(failure " + parse.Quote(e.Failure) + ")"
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

func (e Exitus) String() string {
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

func (e Exitus) Bool() bool {
	return e.Sort == Ok
}

func allok(es []Exitus) bool {
	for _, e := range es {
		if e.Sort != Ok {
			return false
		}
	}
	return true
}

// list is a list of Value's.
type list []Value

func newList() *list {
	l := list([]Value{})
	return &l
}

func (l *list) Type() Type {
	return listType
}

func (l *list) appendStrings(ss []string) {
	for _, s := range ss {
		*l = append(*l, str(s))
	}
}

func (l *list) Repr() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	for i, v := range *l {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(v.Repr())
	}
	buf.WriteRune(']')
	return buf.String()
}

var indexOutOfRange = errors.New("index out of range")

func (l *list) Index(idx string) (Value, error) {
	i, err := strconv.Atoi(idx)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i += len(*l)
	}
	if i < 0 || i >= len(*l) {
		return nil, indexOutOfRange
	}
	return (*l)[i], nil
}

// map_ is a map from string to Value.
// TODO(xiaq): support Value keys.
type map_ map[string]Value

func newMap() map_ {
	return map_(make(map[string]Value))
}

func (m map_) Type() Type {
	return mapType
}

func (m map_) Repr() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	for k, v := range m {
		if buf.Len() > 1 {
			buf.WriteByte(' ')
		}
		buf.WriteByte('&')
		buf.WriteString(parse.Quote(k))
		buf.WriteByte(' ')
		buf.WriteString(v.Repr())
	}
	buf.WriteRune(']')
	return buf.String()
}

func (m map_) Index(idx string) (Value, error) {
	v, ok := m[idx]
	if !ok {
		return nil, errors.New("no such key: " + idx)
	}
	return v, nil
}

// Callable represents Value's that may be called.
type callable interface {
	Value
	Call(ec *evalCtx, args []Value) Exitus
}

// closure is a closure.
type closure struct {
	ArgNames []string
	Op       exitusOp
	Captured map[string]Variable
}

func (c *closure) Type() Type {
	return callableType
}

func newClosure(a []string, op exitusOp, e map[string]Variable) *closure {
	return &closure{a, op, e}
}

func (c *closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

type builtinFn struct {
	Name string
	Impl func(*evalCtx, []Value) Exitus
}

func (b *builtinFn) Type() Type {
	return callableType
}

func (b *builtinFn) Repr() string {
	return "$" + FnPrefix + b.Name
}

type externalCmd struct {
	Name string
}

func (e externalCmd) Type() Type {
	return callableType
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
	return ratType
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

func evalSubscript(ec *evalCtx, l, r Value, lp, rp int) Value {
	left, ok := l.(indexer)
	if !ok {
		ec.errorf(lp, "%s value cannot be indexed", l.Type())
	}

	right, ok := r.(str)
	if !ok {
		ec.errorf(rp, "%s invalid cannot be used as index", r.Type())
	}

	v, err := left.Index(string(right))
	if err != nil {
		ec.errorf(lp, "%v", err)
	}
	return v
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
		vs := list(make([]Value, len(a)))
		for i, v := range a {
			vs[i] = fromJSONInterface(v)
		}
		return &vs
	case map[string]interface{}:
		m := v.(map[string]interface{})
		m_ := newMap()
		for k, v := range m {
			m_[k] = fromJSONInterface(v)
		}
		return m_
	default:
		// TODO Find a better way to report error
		return NewFailure(fmt.Sprintf("unexpected json type: %T", v))
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

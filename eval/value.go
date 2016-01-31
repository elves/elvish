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
	Reprer
}

type Reprer interface {
	Repr() string
}

// Booler represents a Value with a custom semantics of truthness. If a Value
// does not satisfy this interface, it is automatically true.
type Booler interface {
	Bool() bool
}

// Stringer represents a Value with a custom string representation. If a Value
// does not satisfy this interface, its Repr method is used.
type Stringer interface {
	String() string
}

// Indexer represents a Value that may be indexed.
type Indexer interface {
	Index(idx string) (Value, error)
}

// Caller represents a Value that may be called.
type Caller interface {
	Call(ec *evalCtx, args []Value) Error
}

// Type is the type of a value.
type Type int

const (
	TInvalid Type = iota
	TString
	TError
	TBool
	TList
	TMap
	TFn
	TRat
)

// Error definitions.
var (
	needIntIndex    = errors.New("need integer index")
	indexOutOfRange = errors.New("index out of range")
	errOnlyStrOrRat = errors.New("only str or rat may be converted to rat")
)

// String is just a string.
type String string

func (s String) Type() Type {
	return TString
}

func (s String) Repr() string {
	return parse.Quote(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Index(idx string) (Value, error) {
	i, err := strconv.Atoi(idx)
	if err != nil {
		return nil, err
	}

	r, err := strutil.NthRune(string(s), i)
	if err != nil {
		return nil, err
	}
	return String(string(r)), nil
}

// Bool represents truthness.
type Bool bool

func (b Bool) Type() Type {
	return TBool
}

func (b Bool) Repr() string {
	if b {
		return "$true"
	}
	return "$false"
}

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b Bool) Bool() bool {
	return bool(b)
}

// Error represents runtime errors in elvish constructs.
type Error struct {
	inner error
}

func (e Error) Type() Type {
	return TError
}

func (e Error) Repr() string {
	if e.inner == nil {
		return "$ok"
	}
	if r, ok := e.inner.(Reprer); ok {
		return r.Repr()
	}
	return "?(error " + parse.Quote(e.inner.Error()) + ")"
}

func (e Error) String() string {
	if e.inner == nil {
		return "ok"
	}
	return e.inner.Error()
}

func (e Error) Bool() bool {
	return e.inner == nil
}

// Common Error values.
var (
	OK             = Error{nil}
	GenericFailure = Error{errors.New("generic failure")}
)

func NewFailure(text string) Error {
	return Error{errors.New(text)}
}

// multiError is multiple errors packed into one. It is used for reporting
// errors of pipelines, in which multiple forms may error.
type multiError struct {
	errors []Error
}

func (me multiError) Repr() string {
	b := new(bytes.Buffer)
	b.WriteString("(multi-error")
	for _, e := range me.errors {
		b.WriteString(" ")
		b.WriteString(e.Repr())
	}
	b.WriteString(")")
	return b.String()
}

func (me multiError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range me.errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(e.inner.Error())
	}
	b.WriteString(")")
	return b.String()
}

func newMultiError(es ...Error) Error {
	return Error{multiError{es}}
}

// Flow is a special type of Error used for control flows.
type flow uint

const (
	Return flow = iota
	Break
	Continue
)

var flowNames = [...]string{
	"return", "break", "continue",
}

func (f flow) Repr() string {
	return "?(" + f.Error() + ")"
}

func (f flow) Error() string {
	if f >= flow(len(flowNames)) {
		return fmt.Sprintf("!(BAD FLOW: %v)", f)
	}
	return flowNames[f]
}

func newFlow(f flow) Error {
	return Error{f}
}

func allok(es []Error) bool {
	for _, e := range es {
		if e.inner != nil {
			return false
		}
	}
	return true
}

// List is a list of Value's.
type List struct {
	inner *[]Value
}

func NewList(vs ...Value) List {
	return List{&vs}
}

func (l List) Type() Type {
	return TList
}

func (l List) appendStrings(ss []string) {
	for _, s := range ss {
		*l.inner = append(*l.inner, String(s))
	}
}

func (l List) Repr() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	for i, v := range *l.inner {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(v.Repr())
	}
	buf.WriteRune(']')
	return buf.String()
}

func (l List) Index(idx string) (Value, error) {
	i, err := strconv.Atoi(idx)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i += len(*l.inner)
	}
	if i < 0 || i >= len(*l.inner) {
		return nil, indexOutOfRange
	}
	return (*l.inner)[i], nil
}

// Map is a map from string to Value.
// TODO(xiaq): support Value keys.
type Map map[string]Value

func NewMap() Map {
	return Map(make(map[string]Value))
}

func (m Map) Type() Type {
	return TMap
}

func (m Map) Repr() string {
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

func (m Map) Index(idx string) (Value, error) {
	v, ok := m[idx]
	if !ok {
		return nil, errors.New("no such key: " + idx)
	}
	return v, nil
}

// Closure is a closure.
type Closure struct {
	ArgNames []string
	Op       exitusOp
	Captured map[string]Variable
}

func (c *Closure) Type() Type {
	return TFn
}

func newClosure(a []string, op exitusOp, e map[string]Variable) *Closure {
	return &Closure{a, op, e}
}

func (c *Closure) Repr() string {
	return fmt.Sprintf("<Closure%v>", *c)
}

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl func(*evalCtx, []Value) Error
}

func (b *BuiltinFn) Type() Type {
	return TFn
}

func (b *BuiltinFn) Repr() string {
	return "$" + FnPrefix + b.Name
}

// ExternalCmd is an external command.
type ExternalCmd struct {
	Name string
}

func (e ExternalCmd) Type() Type {
	return TFn
}

func (e ExternalCmd) Repr() string {
	return "<external " + e.Name + " >"
}

// Rat is a rational number
type Rat struct {
	b *big.Rat
}

func (r Rat) Type() Type {
	return TRat
}

func (r Rat) Repr() string {
	return "(rat " + r.String() + ")"
}

func (r Rat) String() string {
	if r.b.IsInt() {
		return r.b.Num().String()
	}
	return r.b.String()
}

func evalIndex(ec *evalCtx, l, r Value, lp, rp int) Value {
	left, ok := l.(Indexer)
	if !ok {
		ec.errorf(lp, "%s value cannot be indexed", l.Type())
	}

	right, ok := r.(String)
	if !ok {
		ec.errorf(rp, "%s invalid cannot be used as index", r.Type())
	}

	v, err := left.Index(string(right))
	if err != nil {
		ec.errorf(lp, "%v", err)
	}
	return v
}

// FromJSONInterface converts a interface{} that results from json.Unmarshal to
// a Value.
func FromJSONInterface(v interface{}) Value {
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
		vs := make([]Value, len(a))
		for i, v := range a {
			vs[i] = FromJSONInterface(v)
		}
		return List{&vs}
	case map[string]interface{}:
		m := v.(map[string]interface{})
		m_ := NewMap()
		for k, v := range m {
			m_[k] = FromJSONInterface(v)
		}
		return m_
	default:
		// TODO Find a better way to report error
		return NewFailure(fmt.Sprintf("unexpected json type: %T", v))
	}
}

// Eq compares two Value's.
func Eq(a, b Value) bool {
	// BUG(xiaq): valueEq uses reflect.DeepEqual to check the equality of two
	// values, may can become wrong when values get more complex.
	return reflect.DeepEqual(a, b)
}

// ToString converts a Value to String. When the Value type implements
// String(), it is used. Otherwise Repr() is used.
func ToString(v Value) string {
	if s, ok := v.(Stringer); ok {
		return s.String()
	}
	return v.Repr()
}

// ToBool converts a Value to bool. When the Value type implements Bool(), it
// is used. Otherwise it is considered true.
func ToBool(v Value) bool {
	if b, ok := v.(Booler); ok {
		return b.Bool()
	}
	return true
}

func allTrue(vs []Value) bool {
	for _, v := range vs {
		if !ToBool(v) {
			return false
		}
	}
	return true
}

// ToRat converts a Value to rat. A str can be converted to a rat if it can be
// parsed. A rat is returned as-is. Other types of values cannot be converted.
func ToRat(v Value) (Rat, error) {
	switch v := v.(type) {
	case Rat:
		return v, nil
	case String:
		r := big.Rat{}
		_, err := fmt.Sscanln(string(v), &r)
		if err != nil {
			return Rat{}, fmt.Errorf("%s cannot be parsed as rat", v.Repr())
		}
		return Rat{&r}, nil
	default:
		return Rat{}, errOnlyStrOrRat
	}
}

package eval

import (
	"errors"
	"unicode/utf8"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

// String is just a string.
type String string

var (
	_ types.Value = String("")
	_ ListLike    = String("")
)

var ErrReplacementMustBeString = errors.New("replacement must be string")

func (String) Kind() string {
	return "string"
}

func (s String) Repr(int) string {
	return parse.Quote(string(s))
}

func (s String) Equal(rhs interface{}) bool {
	return s == rhs
}

func (s String) Hash() uint32 {
	return hash.String(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Len() int {
	return len(string(s))
}

func (s String) IndexOne(idx types.Value) types.Value {
	i, j := s.index(idx)
	return s[i:j]
}

func (s String) Assoc(idx, v types.Value) types.Value {
	i, j := s.index(idx)
	repl, ok := v.(String)
	if !ok {
		throw(ErrReplacementMustBeString)
	}
	return s[:i] + repl + s[j:]
}

func (s String) index(idx types.Value) (int, int) {
	slice, i, j := ParseAndFixListIndex(ToString(idx), len(s))
	r, size := utf8.DecodeRuneInString(string(s[i:]))
	if r == utf8.RuneError {
		throw(ErrBadIndex)
	}
	if slice {
		if r, _ := utf8.DecodeLastRuneInString(string(s[:j])); r == utf8.RuneError {
			throw(ErrBadIndex)
		}
		return i, j
	}
	return i, i + size
}

func (s String) Iterate(f func(v types.Value) bool) {
	for _, r := range s {
		b := f(String(string(r)))
		if !b {
			break
		}
	}
}

// Call resolves a command name to either a Fn variable or external command and
// calls it.
func (s String) Call(ec *Frame, args []types.Value, opts map[string]types.Value) {
	resolve(string(s), ec).Call(ec, args, opts)
}

func resolve(s string, ec *Frame) Fn {
	// Try variable
	explode, ns, name := ParseVariable(string(s))
	if !explode {
		if v := ec.ResolveVar(ns, name+FnSuffix); v != nil {
			if caller, ok := v.Get().(Fn); ok {
				return caller
			}
		}
	}

	// External command
	return ExternalCmd{string(s)}
}

// ToString converts a Value to String. When the Value type implements
// String(), it is used. Otherwise Repr(NoPretty) is used.
func ToString(v types.Value) string {
	if s, ok := v.(types.Stringer); ok {
		return s.String()
	}
	return v.Repr(types.NoPretty)
}

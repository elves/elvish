package eval

import (
	"unicode/utf8"

	"github.com/elves/elvish/parse"
)

// String is just a string.
type String string

var (
	_ Value    = String("")
	_ ListLike = String("")
)

func (String) Kind() string {
	return "string"
}

func (s String) Repr(int) string {
	return quote(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Len() int {
	return len(string(s))
}

func (s String) IndexOne(idx Value) Value {
	slice, i, j := ParseAndFixListIndex(ToString(idx), len(s))
	var r rune
	if r, _ = utf8.DecodeRuneInString(string(s[i:])); r == utf8.RuneError {
		throw(ErrBadIndex)
	}
	if slice {
		if r, _ := utf8.DecodeLastRuneInString(string(s[:j])); r == utf8.RuneError {
			throw(ErrBadIndex)
		}
		return String(s[i:j])
	}
	return String(r)
}

func (s String) Iterate(f func(v Value) bool) {
	for _, r := range s {
		b := f(String(string(r)))
		if !b {
			break
		}
	}
}

// Call resolves a command name to either a Fn variable or external command and
// calls it.
func (s String) Call(ec *EvalCtx, args []Value, opts map[string]Value) {
	resolve(string(s), ec).Call(ec, args, opts)
}

func resolve(s string, ec *EvalCtx) FnValue {
	// Try variable
	explode, ns, name := ParseAndFixVariable(string(s))
	if !explode {
		if v := ec.ResolveVar(ns, FnPrefix+name); v != nil {
			if caller, ok := v.Get().(FnValue); ok {
				return caller
			}
		}
	}

	// External command
	return ExternalCmd{string(s)}
}

// ToString converts a Value to String. When the Value type implements
// String(), it is used. Otherwise Repr(NoPretty) is used.
func ToString(v Value) string {
	if s, ok := v.(Stringer); ok {
		return s.String()
	}
	return v.Repr(NoPretty)
}

func quote(s string) string {
	return parse.Quote(s)
}

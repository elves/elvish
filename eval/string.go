package eval

import "github.com/elves/elvish/parse"

// String is just a string.
type String string

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

// Call resolves a command name to either a Fn variable or external command and
// calls it.
func (s String) Call(ec *EvalCtx, args []Value, opts map[string]Value) {
	resolve(string(s), ec).Call(ec, args, opts)
}

func resolve(s string, ec *EvalCtx) FnValue {
	// Try variable
	splice, ns, name := ParseVariable(string(s))
	if !splice {
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

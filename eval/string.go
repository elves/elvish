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

// Call resolves a command name to either a Caller variable or external command
// and calls it.
func (s String) Call(ec *EvalCtx, args []Value) {
	resolve(string(s), ec).Call(ec, args)
}

func resolve(s string, ec *EvalCtx) Caller {
	// Try variable
	splice, ns, name := parseVariable(string(s))
	if !splice {
		if v := ec.ResolveVar(ns, FnPrefix+name); v != nil {
			if caller, ok := v.Get().(Caller); ok {
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

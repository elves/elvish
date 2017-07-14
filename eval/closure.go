package eval

import (
	"errors"
	"fmt"
)

// ErrArityMismatch is thrown by a closure when the number of arguments the user
// supplies does not match with what is required.
var ErrArityMismatch = errors.New("arity mismatch")

var unnamedRestArg = "@"

// Closure is a closure defined in elvish script.
type Closure struct {
	ArgNames []string
	// The name for the rest argument. If empty, the function has fixed arity.
	// If equal to unnamedRestArg, the rest argument is unnamed but can be
	// accessed via $args.
	RestArg    string
	Op         Op
	Captured   map[string]Variable
	SourceName string
	Source     string
}

var _ CallableValue = &Closure{}

// Kind returns "fn".
func (*Closure) Kind() string {
	return "fn"
}

// Eq compares by identity.
func (c *Closure) Eq(rhs interface{}) bool {
	return c == rhs
}

// Repr returns an opaque representation "<closure 0x23333333>".
func (c *Closure) Repr(int) string {
	return fmt.Sprintf("<closure %p>", c)
}

// Call calls a closure.
func (c *Closure) Call(ec *EvalCtx, args []Value, opts map[string]Value) {
	// TODO Support keyword arguments
	if c.RestArg != "" {
		if len(c.ArgNames) > len(args) {
			throw(ErrArityMismatch)
		}
	} else {
		if len(c.ArgNames) != len(args) {
			throw(ErrArityMismatch)
		}
	}

	// This evalCtx is dedicated to the current form, so we modify it in place.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make upvalue namespace and capture variables.
	ec.up = make(map[string]Variable)
	for name, variable := range c.Captured {
		ec.up[name] = variable
	}
	// Make local namespace and pass arguments.
	ec.local = make(map[string]Variable)
	for i, name := range c.ArgNames {
		ec.local[name] = NewPtrVariable(args[i])
	}
	if c.RestArg != "" && c.RestArg != unnamedRestArg {
		ec.local[c.RestArg] = NewPtrVariable(NewList(args[len(c.ArgNames):]...))
	}
	// Logger.Printf("EvalCtx=%p, args=%v, opts=%v", ec, args, opts)
	ec.positionals = args
	ec.local["args"] = NewPtrVariable(NewList(args...))
	// XXX This conversion was done by the other direction.
	convertedOpts := make(map[Value]Value)
	for k, v := range opts {
		convertedOpts[String(k)] = v
	}
	ec.local["opts"] = NewPtrVariable(Map{&convertedOpts})

	ec.traceback = ec.addTraceback()

	ec.srcName, ec.src = c.SourceName, c.Source
	c.Op.Exec(ec)
}

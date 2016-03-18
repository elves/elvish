package eval

import (
	"errors"
	"fmt"
)

var ErrArityMismatch = errors.New("arity mismatch")

var unnamedRestArg = "@"

// Closure is a closure defined in elvish script.
type Closure struct {
	ArgNames []string
	// The name for the rest argument. If empty, the function has fixed arity.
	// If equal to unnamedRestArg, the rest argument is unnamed but can be
	// accessed via $args.
	RestArg  string
	Op       Op
	Captured map[string]Variable
}

func (*Closure) Kind() string {
	return "fn"
}

func newClosure(a []string, r string, op Op, e map[string]Variable) *Closure {
	return &Closure{a, r, op, e}
}

func (c *Closure) Repr(int) string {
	return fmt.Sprintf("<closure%v>", *c)
}

// Call calls a closure.
func (c *Closure) Call(ec *EvalCtx, args []Value) {
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
	Logger.Printf("EvalCtx=%p, args=%v", ec, args)
	ec.positionals = args
	ec.local["args"] = NewPtrVariable(List{&args})
	ec.local["kwargs"] = NewPtrVariable(Map{&map[Value]Value{}})

	// TODO(xiaq): Also change ec.name and ec.text since the closure being
	// called can come from another source.

	c.Op.Exec(ec)
}

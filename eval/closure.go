package eval

import (
	"errors"
	"fmt"
)

var ErrArityMismatch = errors.New("arity mismatch")

// Closure is a closure.
type Closure struct {
	ArgNames []string
	Op       Op
	Captured map[string]Variable
	Variadic bool
}

func (*Closure) Kind() string {
	return "fn"
}

func newClosure(a []string, op Op, e map[string]Variable, v bool) *Closure {
	return &Closure{a, op, e, v}
}

func (c *Closure) Repr(int) string {
	return fmt.Sprintf("<Closure%v>", *c)
}

// Call calls a closure.
func (c *Closure) Call(ec *EvalCtx, args []Value) {
	// TODO Support optional/rest argument
	// TODO Support keyword arguments
	if !c.Variadic && len(args) != len(c.ArgNames) {
		throw(ErrArityMismatch)
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
	if !c.Variadic {
		for i, name := range c.ArgNames {
			ec.local[name] = NewPtrVariable(args[i])
		}
	}
	ec.local["args"] = NewPtrVariable(List{&args})
	ec.local["kwargs"] = NewPtrVariable(Map{&map[Value]Value{}})

	// TODO(xiaq): Also change ec.name and ec.text since the closure being
	// called can come from another source.

	c.Op.Exec(ec)
}

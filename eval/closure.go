package eval

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/xiaq/persistent/hash"
)

// ErrArityMismatch is thrown by a closure when the number of arguments the user
// supplies does not match with what is required.
var ErrArityMismatch = errors.New("arity mismatch")

var unnamedRestArg = "@"

// Closure is a closure defined in elvish script.
type Closure struct {
	ArgNames []string
	// The name for the rest argument. If empty, the function has fixed arity.
	RestArg     string
	OptNames    []string
	OptDefaults []Value
	Op          Op
	Captured    map[string]Variable
	SourceName  string
	Source      string
}

var _ CallableValue = &Closure{}

// Kind returns "fn".
func (*Closure) Kind() string {
	return "fn"
}

// Equal compares by identity.
func (c *Closure) Equal(rhs interface{}) bool {
	return c == rhs
}

func (c *Closure) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(c))
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
	if c.RestArg != "" {
		ec.local[c.RestArg] = NewPtrVariable(NewList(args[len(c.ArgNames):]...))
	}
	// Logger.Printf("EvalCtx=%p, args=%v, opts=%v", ec, args, opts)
	for i, name := range c.OptNames {
		v, ok := opts[name]
		if !ok {
			v = c.OptDefaults[i]
		}
		ec.local[name] = NewPtrVariable(v)
	}
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

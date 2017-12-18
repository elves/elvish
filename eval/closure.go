package eval

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/xiaq/persistent/hash"
	"github.com/xiaq/persistent/hashmap"
)

// ErrArityMismatch is thrown by a closure when the number of arguments the user
// supplies does not match with what is required.
var ErrArityMismatch = errors.New("arity mismatch")

// Closure is a closure defined in elvish script.
type Closure struct {
	ArgNames []string
	// The name for the rest argument. If empty, the function has fixed arity.
	RestArg     string
	OptNames    []string
	OptDefaults []Value
	Op          Op
	Captured    Scope
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
	// TODO(xiaq): Is it safe to simply assign ec.up = c.Captured?
	ec.up = makeScope()
	for name, variable := range c.Captured.Names {
		ec.up.Names[name] = variable
	}
	for name, use := range c.Captured.Uses {
		ec.up.Uses[name] = use
	}

	// Populate local scope with arguments, possibly a rest argument, and
	// options.
	ec.local = makeScope()
	for i, name := range c.ArgNames {
		ec.local.Names[name] = NewPtrVariable(args[i])
	}
	if c.RestArg != "" {
		ec.local.Names[c.RestArg] = NewPtrVariable(NewList(args[len(c.ArgNames):]...))
	}
	// Logger.Printf("EvalCtx=%p, args=%v, opts=%v", ec, args, opts)
	for i, name := range c.OptNames {
		v, ok := opts[name]
		if !ok {
			v = c.OptDefaults[i]
		}
		ec.local.Names[name] = NewPtrVariable(v)
	}
	// XXX This conversion was done by the other direction.
	convertedOpts := hashmap.Empty
	for k, v := range opts {
		convertedOpts = convertedOpts.Assoc(String(k), v)
	}
	ec.local.Names["opts"] = NewPtrVariable(NewMap(convertedOpts))

	ec.traceback = ec.addTraceback()

	ec.srcName, ec.src = c.SourceName, c.Source
	c.Op.Exec(ec)
}

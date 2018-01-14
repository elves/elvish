package eval

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
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
	OptDefaults []types.Value
	Op          Op
	Captured    Ns
	SrcMeta     *Source
}

var _ Fn = &Closure{}

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
func (c *Closure) Call(ec *Frame, args []types.Value, opts map[string]types.Value) {
	if c.RestArg != "" {
		if len(c.ArgNames) > len(args) {
			throwf("need %d or more arguments, got %d", len(c.ArgNames), len(args))
		}
	} else {
		if len(c.ArgNames) != len(args) {
			throwf("need %d arguments, got %d", len(c.ArgNames), len(args))
		}
	}

	// This evalCtx is dedicated to the current form, so we modify it in place.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make upvalue namespace and capture variables.
	// TODO(xiaq): Is it safe to simply assign ec.up = c.Captured?
	ec.up = make(Ns)
	for name, variable := range c.Captured {
		ec.up[name] = variable
	}

	// Populate local scope with arguments, possibly a rest argument, and
	// options.
	ec.local = make(Ns)
	for i, name := range c.ArgNames {
		ec.local[name] = vartypes.NewPtr(args[i])
	}
	if c.RestArg != "" {
		ec.local[c.RestArg] = vartypes.NewPtr(types.MakeList(args[len(c.ArgNames):]...))
	}
	optUsed := make(map[string]struct{})
	for i, name := range c.OptNames {
		v, ok := opts[name]
		if ok {
			optUsed[name] = struct{}{}
		} else {
			v = c.OptDefaults[i]
		}
		ec.local[name] = vartypes.NewPtr(v)
	}
	for name := range opts {
		_, used := optUsed[name]
		if !used {
			throwf("Unknown option %s", parse.Quote(name))
		}
	}

	ec.traceback = ec.addTraceback()

	ec.srcMeta = c.SrcMeta
	c.Op.Exec(ec)
}

package eval

import (
	"fmt"
	"unsafe"

	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/hash"
)

// Closure is a closure defined in Elvish script. Each closure has its unique
// identity.
type Closure struct {
	ArgNames []string
	// The name for the rest argument. If empty, the function has fixed arity.
	RestArg     string
	OptNames    []string
	OptDefaults []interface{}
	Op          effectOp
	Captured    Ns
	SrcMeta     parse.Source
	DefFrom     int
	DefTo       int
}

var _ Callable = &Closure{}

// Kind returns "fn".
func (*Closure) Kind() string {
	return "fn"
}

// Equal compares by address.
func (c *Closure) Equal(rhs interface{}) bool {
	return c == rhs
}

// Hash returns the hash of the address of the closure.
func (c *Closure) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(c))
}

// Repr returns an opaque representation "<closure 0x23333333>".
func (c *Closure) Repr(int) string {
	return fmt.Sprintf("<closure %p>", c)
}

func listOfStrings(ss []string) vals.List {
	list := vals.EmptyList
	for _, s := range ss {
		list = list.Cons(s)
	}
	return list
}

// Call calls a closure.
func (c *Closure) Call(fm *Frame, args []interface{}, opts map[string]interface{}) error {
	if c.RestArg != "" {
		if len(args) < len(c.ArgNames) {
			return errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: len(c.ArgNames), ValidHigh: -1, Actual: len(args)}
		}
	} else {
		if len(args) != len(c.ArgNames) {
			return errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: len(c.ArgNames), ValidHigh: len(c.ArgNames), Actual: len(args)}
		}
	}

	// This evalCtx is dedicated to the current form, so we modify it in place.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make upvalue namespace and capture variables.
	// TODO(xiaq): Is it safe to simply assign ec.up = c.Captured?
	fm.up = make(Ns)
	for name, variable := range c.Captured {
		fm.up[name] = variable
	}

	// Populate local scope with arguments, possibly a rest argument, and
	// options.
	fm.local = make(Ns)
	for i, name := range c.ArgNames {
		fm.local[name] = vars.FromInit(args[i])
	}
	if c.RestArg != "" {
		fm.local[c.RestArg] = vars.FromInit(vals.MakeList(args[len(c.ArgNames):]...))
	}
	optUsed := make(map[string]struct{})
	for i, name := range c.OptNames {
		v, ok := opts[name]
		if ok {
			optUsed[name] = struct{}{}
		} else {
			v = c.OptDefaults[i]
		}
		fm.local[name] = vars.FromInit(v)
	}
	for name := range opts {
		_, used := optUsed[name]
		if !used {
			return fmt.Errorf("unknown option %s", parse.Quote(name))
		}
	}

	fm.srcMeta = c.SrcMeta
	return c.Op.exec(fm)
}

func (c *Closure) Fields() vals.StructMap { return closureFields{c} }

type closureFields struct{ c *Closure }

func (closureFields) IsStructMap() {}

func (cf closureFields) ArgNames() vals.List { return listOfStrings(cf.c.ArgNames) }
func (cf closureFields) RestArg() string     { return cf.c.RestArg }
func (cf closureFields) OptNames() vals.List { return listOfStrings(cf.c.OptNames) }
func (cf closureFields) Src() parse.Source   { return cf.c.SrcMeta }

func (cf closureFields) OptDefaults() vals.List {
	return vals.MakeList(cf.c.OptDefaults...)
}

func (cf closureFields) Body() string {
	return cf.c.SrcMeta.Code[cf.c.Op.From:cf.c.Op.To]
}

func (cf closureFields) Def() string {
	return cf.c.SrcMeta.Code[cf.c.DefFrom:cf.c.DefTo]
}

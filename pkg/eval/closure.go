package eval

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/hash"
)

// A user-defined function in Elvish code. Each closure has its unique identity.
type closure struct {
	ArgNames []string
	// The index of the rest argument. -1 if there is no rest argument.
	RestArg     int
	OptNames    []string
	OptDefaults []interface{}
	Op          effectOp
	Captured    Ns
	SrcMeta     parse.Source
	DefRange    diag.Ranging
}

var _ Callable = &closure{}

// Kind returns "fn".
func (*closure) Kind() string {
	return "fn"
}

// Equal compares by address.
func (c *closure) Equal(rhs interface{}) bool {
	return c == rhs
}

// Hash returns the hash of the address of the closure.
func (c *closure) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(c))
}

// Repr returns an opaque representation "<closure 0x23333333>".
func (c *closure) Repr(int) string {
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
func (c *closure) Call(fm *Frame, args []interface{}, opts map[string]interface{}) error {
	if c.RestArg != -1 {
		if len(args) < len(c.ArgNames)-1 {
			return errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: len(c.ArgNames) - 1, ValidHigh: -1, Actual: len(args)}
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

	// Populate local scope with arguments and options.
	fm.local = make(Ns)
	if c.RestArg == -1 {
		for i, name := range c.ArgNames {
			fm.local[name] = vars.FromInit(args[i])
		}
	} else {
		for i := 0; i < c.RestArg; i++ {
			fm.local[c.ArgNames[i]] = vars.FromInit(args[i])
		}
		restOff := len(args) - len(c.ArgNames)
		fm.local[c.ArgNames[c.RestArg]] = vars.FromInit(
			vals.MakeList(args[c.RestArg : c.RestArg+restOff+1]...))
		for i := c.RestArg + 1; i < len(c.ArgNames); i++ {
			fm.local[c.ArgNames[i]] = vars.FromInit(args[i+restOff])
		}
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

func (c *closure) Fields() vals.StructMap { return closureFields{c} }

type closureFields struct{ c *closure }

func (closureFields) IsStructMap() {}

func (cf closureFields) ArgNames() vals.List { return listOfStrings(cf.c.ArgNames) }
func (cf closureFields) RestArg() string     { return strconv.Itoa(cf.c.RestArg) }
func (cf closureFields) OptNames() vals.List { return listOfStrings(cf.c.OptNames) }
func (cf closureFields) Src() parse.Source   { return cf.c.SrcMeta }

func (cf closureFields) OptDefaults() vals.List {
	return vals.MakeList(cf.c.OptDefaults...)
}

func (cf closureFields) Body() string {
	r := cf.c.Op.(diag.Ranger).Range()
	return cf.c.SrcMeta.Code[r.From:r.To]
}

func (cf closureFields) Def() string {
	return cf.c.SrcMeta.Code[cf.c.DefRange.From:cf.c.DefRange.To]
}

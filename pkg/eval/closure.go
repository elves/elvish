package eval

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/xiaq/persistent/hash"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

// A user-defined function in Elvish code. Each closure has its unique identity.
type closure struct {
	ArgNames []string
	// The index of the rest argument. -1 if there is no rest argument.
	RestArg     int
	OptNames    []string
	OptDefaults []interface{}
	Op          effectOp
	NewLocal    []string
	Captured    *Ns
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

// Call calls a closure.
func (c *closure) Call(fm *Frame, args []interface{}, opts map[string]interface{}) error {
	// Check number of arguments.
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
	// Check whether all supplied options are supported. This map contains the
	// subset of keys from opts that can be found in c.OptNames.
	optSupported := make(map[string]struct{})
	for _, name := range c.OptNames {
		_, ok := opts[name]
		if ok {
			optSupported[name] = struct{}{}
		}
	}
	if len(optSupported) < len(opts) {
		// Report all the options that are not supported.
		unsupported := make([]string, 0, len(opts)-len(optSupported))
		for name := range opts {
			_, supported := optSupported[name]
			if !supported {
				unsupported = append(unsupported, parse.Quote(name))
			}
		}
		sort.Strings(unsupported)
		return UnsupportedOptionsError{unsupported}
	}

	// This Frame is dedicated to the current form, so we can modify it in place.

	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make upvalue namespace and capture variables.
	fm.up = c.Captured

	// Populate local scope with arguments, options, and newly created locals.
	localSize := len(c.ArgNames) + len(c.OptNames) + len(c.NewLocal)
	local := &Ns{make([]vars.Var, localSize), make([]string, localSize), make([]bool, localSize)}

	for i, name := range c.ArgNames {
		local.names[i] = name
	}
	if c.RestArg == -1 {
		for i := range c.ArgNames {
			local.slots[i] = vars.FromInit(args[i])
		}
	} else {
		for i := 0; i < c.RestArg; i++ {
			local.slots[i] = vars.FromInit(args[i])
		}
		restOff := len(args) - len(c.ArgNames)
		local.slots[c.RestArg] = vars.FromInit(
			vals.MakeList(args[c.RestArg : c.RestArg+restOff+1]...))
		for i := c.RestArg + 1; i < len(c.ArgNames); i++ {
			local.slots[i] = vars.FromInit(args[i+restOff])
		}
	}

	offset := len(c.ArgNames)
	for i, name := range c.OptNames {
		v, ok := opts[name]
		if !ok {
			v = c.OptDefaults[i]
		}
		local.names[offset+i] = name
		local.slots[offset+i] = vars.FromInit(v)
	}

	offset += len(c.OptNames)
	for i, name := range c.NewLocal {
		local.names[offset+i] = name
		local.slots[offset+i] = MakeVarFromName(name)
	}

	fm.local = local
	fm.srcMeta = c.SrcMeta
	return c.Op.exec(fm)
}

// MakeVarFromName creates a Var with a suitable type constraint inferred from
// the name.
func MakeVarFromName(name string) vars.Var {
	switch {
	case strings.HasSuffix(name, FnSuffix):
		val := Callable(nil)
		return vars.FromPtr(&val)
	case strings.HasSuffix(name, NsSuffix):
		val := (*Ns)(nil)
		return vars.FromPtr(&val)
	default:
		return vars.FromInit(nil)
	}
}

// UnsupportedOptionsError is an error returned by a closure call when there are
// unsupported options.
type UnsupportedOptionsError struct {
	Options []string
}

func (er UnsupportedOptionsError) Error() string {
	if len(er.Options) == 1 {
		return fmt.Sprintf("unsupported option: %s", er.Options[0])
	}
	return fmt.Sprintf("unsupported options: %s", strings.Join(er.Options, ", "))
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

func listOfStrings(ss []string) vals.List {
	list := vals.EmptyList
	for _, s := range ss {
		list = list.Cons(s)
	}
	return list
}

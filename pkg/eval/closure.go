package eval

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/persistent/hash"
)

// Closure is a function defined with Elvish code. Each Closure has its unique
// identity.
type Closure struct {
	ArgNames []string
	// The index of the rest argument. -1 if there is no rest argument.
	RestArg     int
	OptNames    []string
	OptDefaults []any
	Src         parse.Source
	DefRange    diag.Ranging
	op          effectOp
	newLocal    []staticVarInfo
	captured    *Ns
}

var (
	_ Callable       = &Closure{}
	_ vals.PseudoMap = &Closure{}
)

// Kind returns "fn".
func (*Closure) Kind() string {
	return "fn"
}

// Equal compares by address.
func (c *Closure) Equal(rhs any) bool {
	return c == rhs
}

// Hash returns the hash of the address of the closure.
func (c *Closure) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(c))
}

// Call calls a closure.
func (c *Closure) Call(fm *Frame, args []any, opts map[string]any) error {
	// Check number of arguments.
	if c.RestArg != -1 {
		if len(args) < len(c.ArgNames)-1 {
			return errs.ArityMismatch{What: "arguments",
				ValidLow: len(c.ArgNames) - 1, ValidHigh: -1, Actual: len(args)}
		}
	} else {
		if len(args) != len(c.ArgNames) {
			return errs.ArityMismatch{What: "arguments",
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
	fm.up = c.captured

	// Populate local scope with arguments, options, and newly created locals.
	localSize := len(c.ArgNames) + len(c.OptNames) + len(c.newLocal)
	local := &Ns{make([]vars.Var, localSize), make([]staticVarInfo, localSize)}

	for i, name := range c.ArgNames {
		local.infos[i] = staticVarInfo{name, false, false}
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
		local.infos[offset+i] = staticVarInfo{name, false, false}
		local.slots[offset+i] = vars.FromInit(v)
	}

	offset += len(c.OptNames)
	for i, info := range c.newLocal {
		local.infos[offset+i] = info
		// TODO: Take info.readOnly into account too when creating variable
		local.slots[offset+i] = MakeVarFromName(info.name)
	}

	fm.local = local
	fm.src = c.Src
	fm.defers = new([]func(*Frame) Exception)
	exc := c.op.exec(fm)
	excDefer := fm.runDefers()
	// TODO: Combine exc and excDefer if both are not nil
	if excDefer != nil && exc == nil {
		exc = excDefer
	}
	return exc
}

// MakeVarFromName creates a Var with a suitable type constraint inferred from
// the name.
func MakeVarFromName(name string) vars.Var {
	switch {
	case strings.HasSuffix(name, FnSuffix):
		val := nopGoFn
		return vars.FromPtr(&val)
	case strings.HasSuffix(name, NsSuffix):
		val := &Ns{}
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

func (c *Closure) Fields() vals.MethodMap { return closureFields{c} }

type closureFields struct{ c *Closure }

func (cf closureFields) ArgNames() vals.List { return vals.MakeListSlice(cf.c.ArgNames) }
func (cf closureFields) RestArg() string     { return strconv.Itoa(cf.c.RestArg) }
func (cf closureFields) OptNames() vals.List { return vals.MakeListSlice(cf.c.OptNames) }
func (cf closureFields) Src() parse.Source   { return cf.c.Src }

func (cf closureFields) OptDefaults() vals.List {
	return vals.MakeList(cf.c.OptDefaults...)
}

func (cf closureFields) Body() string {
	r := cf.c.op.(diag.Ranger).Range()
	return cf.c.Src.Code[r.From:r.To]
}

func (cf closureFields) Def() string {
	return cf.c.Src.Code[cf.c.DefRange.From:cf.c.DefRange.To]
}

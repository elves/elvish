package eval

// Misc builtin functions.

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

var (
	ErrNegativeSleepDuration = errors.New("sleep duration must be >= zero")
	ErrInvalidSleepDuration  = errors.New("invalid sleep duration")
)

// Builtins that have not been put into their own groups go here.

func init() {
	addBuiltinFns(map[string]any{
		"kind-of":    kindOf,
		"constantly": constantly,

		// Introspection
		"call":    call,
		"resolve": resolve,
		"eval":    eval,
		"use-mod": useMod,

		"deprecate": deprecate,

		"-ifaddrs": _ifaddrs,
	})

}

var nopGoFn = NewGoFn("nop", nop)

func nop(opts RawOptions, args ...any) {
	// Do nothing
}

func kindOf(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.Kind(a))
		if err != nil {
			return err
		}
	}
	return nil
}

func constantly(args ...any) Callable {
	// TODO(xiaq): Repr of this function is not right.
	return NewGoFn(
		"created by constantly",
		func(fm *Frame) error {
			out := fm.ValueOutput()
			for _, v := range args {
				err := out.Put(v)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func call(fm *Frame, fn Callable, argsVal vals.List, optsVal vals.Map) error {
	args := make([]any, 0, argsVal.Len())
	for it := argsVal.Iterator(); it.HasElem(); it.Next() {
		args = append(args, it.Elem())
	}
	opts := make(map[string]any, optsVal.Len())
	for it := optsVal.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		ks, ok := k.(string)
		if !ok {
			return errs.BadValue{What: "option key",
				Valid: "string", Actual: vals.Kind(k)}
		}
		opts[ks] = v
	}
	return fn.Call(fm.Fork(), args, opts)
}

func resolve(fm *Frame, head string) string {
	special, fnRef := resolveCmdHeadInternally(fm, head, nil)
	switch {
	case special != nil:
		return "special"
	case fnRef != nil:
		return "$" + head + FnSuffix
	default:
		return "(external " + parse.Quote(head) + ")"
	}
}

type evalOpts struct {
	Ns    *Ns
	OnEnd Callable
}

func (*evalOpts) SetDefaultOptions() {}

func eval(fm *Frame, opts evalOpts, code string) error {
	src := parse.Source{Name: fmt.Sprintf("[eval %d]", nextEvalCount()), Code: code}
	ns := opts.Ns
	if ns == nil {
		ns = CombineNs(fm.up, fm.local)
	}
	// The stacktrace already contains the line that calls "eval", so we pass
	// nil as the second argument.
	newNs, exc := fm.Eval(src, nil, ns)
	if opts.OnEnd != nil {
		newFm := fm.Fork()
		errCb := opts.OnEnd.Call(newFm, []any{newNs}, NoOpts)
		if exc == nil {
			return errCb
		}
	}
	return exc
}

// Used to generate unique names for each source passed to eval.
var (
	evalCount      int
	evalCountMutex sync.Mutex
	nextEvalCount  = nextEvalCountImpl
)

func nextEvalCountImpl() int {
	evalCountMutex.Lock()
	defer evalCountMutex.Unlock()
	evalCount++
	return evalCount
}

func useMod(fm *Frame, spec string) (*Ns, error) {
	return use(fm, spec, nil)
}

func deprecate(fm *Frame, msg string) {
	var ctx *diag.Context
	if fm.traceback.Next != nil {
		ctx = fm.traceback.Next.Head
	}
	fm.Deprecate(msg, ctx, 0)
}

func _ifaddrs(fm *Frame) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	out := fm.ValueOutput()
	for _, addr := range addrs {
		err := out.Put(addr.String())
		if err != nil {
			return err
		}
	}
	return nil
}

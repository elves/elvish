package eval

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/elves/elvish/util"
)

// Sequence, list and maps.

func init() {
	addToBuiltinFns([]*BuiltinFn{
		{"ns", nsFn},

		{"range", rangeFn},
		{"repeat", repeat},
		{"explode", explode},

		{"assoc", assoc},
		{"dissoc", dissoc},

		{"all", all},
		{"take", take},
		{"drop", drop},

		{"has-key", hasKey},
		{"has-value", hasValue},

		{"count", count},

		{"keys", keys},
	})
}

func nsFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	ec.OutputChan() <- make(Ns)
}

func rangeFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	var step float64
	ScanOpts(opts, OptToScan{"step", &step, String("1")})

	var lower, upper float64
	var err error

	switch len(args) {
	case 1:
		upper, err = toFloat(args[0])
		maybeThrow(err)
	case 2:
		lower, err = toFloat(args[0])
		maybeThrow(err)
		upper, err = toFloat(args[1])
		maybeThrow(err)
	default:
		throw(ErrArgs)
	}

	out := ec.ports[1].Chan
	for f := lower; f < upper; f += step {
		out <- floatToString(f)
	}
}

func repeat(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		n int
		v Value
	)
	ScanArgs(args, &n, &v)
	TakeNoOpt(opts)

	out := ec.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(ec *EvalCtx, args []Value, opts map[string]Value) {
	var v IterableValue
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	v.Iterate(func(e Value) bool {
		out <- e
		return true
	})
}

func assoc(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		a    Assocer
		k, v Value
	)
	ScanArgs(args, &a, &k, &v)
	TakeNoOpt(opts)
	ec.OutputChan() <- a.Assoc(k, v)
}

func dissoc(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		a Dissocer
		k Value
	)
	ScanArgs(args, &a, &k)
	TakeNoOpt(opts)
	ec.OutputChan() <- a.Dissoc(k)
}

func all(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	valuesDone := make(chan struct{})
	go func() {
		for input := range ec.ports[0].Chan {
			ec.ports[1].Chan <- input
		}
		close(valuesDone)
	}()
	_, err := io.Copy(ec.ports[1].File, ec.ports[0].File)
	<-valuesDone
	if err != nil {
		throwf("cannot copy byte input: %s", err)
	}
}

func take(ec *EvalCtx, args []Value, opts map[string]Value) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v Value) {
		if i < n {
			out <- v
		}
		i++
	})
}

func drop(ec *EvalCtx, args []Value, opts map[string]Value) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v Value) {
		if i >= n {
			out <- v
		}
		i++
	})
}

func hasValue(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var container, value Value
	var found bool

	ScanArgs(args, &container, &value)

	switch container := container.(type) {
	case Iterable:
		container.Iterate(func(v Value) bool {
			found = (v == value)
			return !found
		})
	case MapLike:
		container.IterateKey(func(v Value) bool {
			found = (container.IndexOne(v) == value)
			return !found
		})
	default:
		throw(fmt.Errorf("argument of type '%s' is not iterable", container.Kind()))
	}

	ec.ports[1].Chan <- Bool(found)
}

func hasKey(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var container, key Value
	var found bool

	ScanArgs(args, &container, &key)

	switch container := container.(type) {
	case HasKeyer:
		found = container.HasKey(key)
	case Lener:
		// XXX(xiaq): Not all types that implement Lener have numerical indices
		err := util.PCall(func() {
			ParseAndFixListIndex(ToString(key), container.Len())
		})
		found = (err == nil)
	default:
		throw(fmt.Errorf("couldn't get key or index of type '%s'", container.Kind()))
	}

	ec.ports[1].Chan <- Bool(found)
}

func count(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var n int
	switch len(args) {
	case 0:
		// Count inputs.
		ec.IterateInputs(func(Value) {
			n++
		})
	case 1:
		// Get length of argument.
		v := args[0]
		if lener, ok := v.(Lener); ok {
			n = lener.Len()
		} else if iterator, ok := v.(Iterable); ok {
			iterator.Iterate(func(Value) bool {
				n++
				return true
			})
		} else {
			throw(fmt.Errorf("cannot get length of a %s", v.Kind()))
		}
	default:
		throw(errors.New("want 0 or 1 argument"))
	}
	ec.ports[1].Chan <- String(strconv.Itoa(n))
}

func keys(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var iter IterateKeyer
	ScanArgs(args, &iter)

	out := ec.ports[1].Chan

	iter.IterateKey(func(v Value) bool {
		out <- v
		return true
	})
}

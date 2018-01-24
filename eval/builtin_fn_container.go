package eval

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/elves/elvish/eval/types"
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

func nsFn(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	ec.OutputChan() <- make(Ns)
}

func rangeFn(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var step float64
	ScanOpts(opts, OptToScan{"step", &step, types.String("1")})

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

func repeat(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		n int
		v types.Value
	)
	ScanArgs(args, &n, &v)
	TakeNoOpt(opts)

	out := ec.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var v types.IteratorValue
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	v.Iterate(func(e types.Value) bool {
		out <- e
		return true
	})
}

func assoc(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		a    types.Assocer
		k, v types.Value
	)
	ScanArgs(args, &a, &k, &v)
	TakeNoOpt(opts)
	result, err := a.Assoc(k, v)
	maybeThrow(err)
	ec.OutputChan() <- result
}

func dissoc(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		a types.Dissocer
		k types.Value
	)
	ScanArgs(args, &a, &k)
	TakeNoOpt(opts)
	ec.OutputChan() <- a.Dissoc(k)
}

func all(ec *Frame, args []types.Value, opts map[string]types.Value) {
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

func take(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v types.Value) {
		if i < n {
			out <- v
		}
		i++
	})
}

func drop(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v types.Value) {
		if i >= n {
			out <- v
		}
		i++
	})
}

func hasValue(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	var container, value types.Value
	var found bool

	ScanArgs(args, &container, &value)

	switch container := container.(type) {
	case types.Iterator:
		container.Iterate(func(v types.Value) bool {
			found = (v == value)
			return !found
		})
	case types.MapLike:
		container.IteratePair(func(_, v types.Value) bool {
			found = v == value
			return !found
		})
	default:
		throw(fmt.Errorf("argument of type '%s' is not iterable", types.Kind(container)))
	}

	ec.ports[1].Chan <- types.Bool(found)
}

func hasKey(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	var container, key types.Value
	var found bool

	ScanArgs(args, &container, &key)

	switch container := container.(type) {
	case types.HasKeyer:
		found = container.HasKey(key)
	case types.Lener:
		// XXX(xiaq): Not all types that implement Lener have numerical indices
		_, _, _, err := types.ParseAndFixListIndex(types.ToString(key), container.Len())
		found = (err == nil)
	default:
		throw(fmt.Errorf("couldn't get key or index of type '%s'", types.Kind(container)))
	}

	ec.ports[1].Chan <- types.Bool(found)
}

func count(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	var n int
	switch len(args) {
	case 0:
		// Count inputs.
		ec.IterateInputs(func(types.Value) {
			n++
		})
	case 1:
		// Get length of argument.
		v := args[0]
		if lener, ok := v.(types.Lener); ok {
			n = lener.Len()
		} else if iterator, ok := v.(types.Iterator); ok {
			iterator.Iterate(func(types.Value) bool {
				n++
				return true
			})
		} else {
			throw(fmt.Errorf("cannot get length of a %s", types.Kind(v)))
		}
	default:
		throw(errors.New("want 0 or 1 argument"))
	}
	ec.ports[1].Chan <- types.String(strconv.Itoa(n))
}

func keys(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	var iter types.IterateKeyer
	ScanArgs(args, &iter)

	out := ec.ports[1].Chan

	iter.IterateKey(func(v types.Value) bool {
		out <- v
		return true
	})
}

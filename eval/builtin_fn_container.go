package eval

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/hashmap"
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

var errKeyMustBeString = errors.New("key must be string")

func nsFn(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var m hashmap.Map
	ScanArgs(args, &m)

	ns := make(Ns)
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			throw(errKeyMustBeString)
		}
		ns[kstring] = vartypes.NewPtr(v)
	}
	ec.OutputChan() <- ns
}

func rangeFn(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var step float64
	ScanOpts(opts, OptToScan{"step", &step, "1"})

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

func repeat(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		n int
		v interface{}
	)
	ScanArgs(args, &n, &v)
	TakeNoOpt(opts)

	out := ec.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var v interface{}
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	err := types.Iterate(v, func(e interface{}) bool {
		out <- e
		return true
	})
	maybeThrow(err)
}

func assoc(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var a, k, v interface{}
	ScanArgs(args, &a, &k, &v)
	TakeNoOpt(opts)
	result, err := types.Assoc(a, k, v)
	maybeThrow(err)
	ec.OutputChan() <- result
}

var errCannotDissoc = errors.New("cannot dissoc")

func dissoc(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var a, k interface{}
	ScanArgs(args, &a, &k)
	TakeNoOpt(opts)
	a2 := types.Dissoc(a, k)
	if a2 == nil {
		throw(errCannotDissoc)
	}
	ec.OutputChan() <- a2
}

func all(ec *Frame, args []interface{}, opts map[string]interface{}) {
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

func take(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v interface{}) {
		if i < n {
			out <- v
		}
		i++
	})
}

func drop(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var n int
	iterate := ScanArgsOptionalInput(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v interface{}) {
		if i >= n {
			out <- v
		}
		i++
	})
}

func hasValue(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var container, value interface{}
	var found bool

	ScanArgs(args, &container, &value)

	switch container := container.(type) {
	case hashmap.Map:
		for it := container.Iterator(); it.HasElem(); it.Next() {
			_, v := it.Elem()
			if types.Equal(v, value) {
				found = true
				break
			}
		}
	default:
		err := types.Iterate(container, func(v interface{}) bool {
			found = (v == value)
			return !found
		})
		maybeThrow(err)
	}

	ec.ports[1].Chan <- types.Bool(found)
}

func hasKey(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var container, key interface{}
	var found bool

	ScanArgs(args, &container, &key)

	switch container := container.(type) {
	case hashmap.Map:
		found = hashmap.HasKey(container, key)
	default:
		if len := types.Len(container); len >= 0 {
			// XXX(xiaq): Not all types that implement Lener have numerical indices
			_, err := types.ConvertListIndex(key, len)
			found = (err == nil)
			break
		}
		throw(fmt.Errorf("couldn't get key or index of type '%s'", types.Kind(container)))
	}

	ec.ports[1].Chan <- types.Bool(found)
}

func count(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var n int
	switch len(args) {
	case 0:
		// Count inputs.
		ec.IterateInputs(func(interface{}) {
			n++
		})
	case 1:
		// Get length of argument.
		v := args[0]
		if len := types.Len(v); len >= 0 {
			n = len
		} else {
			err := types.Iterate(v, func(interface{}) bool {
				n++
				return true
			})
			if err != nil {
				throw(fmt.Errorf("cannot get length of a %s", types.Kind(v)))
			}
		}
	default:
		throw(errors.New("want 0 or 1 argument"))
	}
	ec.ports[1].Chan <- strconv.Itoa(n)
}

func keys(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var m hashmap.Map
	ScanArgs(args, &m)

	out := ec.ports[1].Chan

	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		out <- k
	}
}

package eval

import (
	"errors"
	"fmt"
	"io"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

// Sequence, list and maps.

func init() {
	addBuiltinFns(map[string]interface{}{
		"ns": nsFn,

		"range":   rangeFn,
		"repeat":  repeat,
		"explode": explode,

		"assoc":  assoc,
		"dissoc": dissoc,

		"all": all,

		"has-key":   hasKey,
		"has-value": hasValue,

		"take":  take,
		"drop":  drop,
		"count": count,

		"keys": keys,
	})
}

var errKeyMustBeString = errors.New("key must be string")

func nsFn(m hashmap.Map) Ns {
	ns := make(Ns)
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			throw(errKeyMustBeString)
		}
		ns[kstring] = vars.NewAnyWithInit(v)
	}
	return ns
}

func rangeFn(fm *Frame, opts Options, args ...float64) error {
	var step float64
	opts.Scan(OptToScan{"step", &step, "1"})

	var lower, upper float64

	switch len(args) {
	case 1:
		upper = args[0]
	case 2:
		lower, upper = args[0], args[1]
	default:
		return ErrArgs
	}

	out := fm.ports[1].Chan
	for f := lower; f < upper; f += step {
		out <- floatToElv(f)
	}
	return nil
}

func repeat(ec *Frame, n int, v interface{}) {
	out := ec.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(ec *Frame, v interface{}) {
	out := ec.ports[1].Chan
	err := types.Iterate(v, func(e interface{}) bool {
		out <- e
		return true
	})
	maybeThrow(err)
}

func assoc(a, k, v interface{}) (interface{}, error) {
	return types.Assoc(a, k, v)
}

var errCannotDissoc = errors.New("cannot dissoc")

func dissoc(a, k interface{}) (interface{}, error) {
	a2 := types.Dissoc(a, k)
	if a2 == nil {
		return nil, errCannotDissoc
	}
	return a2, nil
}

func all(ec *Frame) error {
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
		return fmt.Errorf("cannot copy byte input: %s", err)
	}
	return nil
}

func take(fm *Frame, n int, inputs Inputs) {
	out := fm.ports[1].Chan
	i := 0
	inputs(func(v interface{}) {
		if i < n {
			out <- v
		}
		i++
	})
}

func drop(fm *Frame, n int, inputs Inputs) {
	out := fm.ports[1].Chan
	i := 0
	inputs(func(v interface{}) {
		if i >= n {
			out <- v
		}
		i++
	})
}

func hasValue(container, value interface{}) (bool, error) {
	switch container := container.(type) {
	case hashmap.Map:
		for it := container.Iterator(); it.HasElem(); it.Next() {
			_, v := it.Elem()
			if types.Equal(v, value) {
				return true, nil
			}
		}
		return false, nil
	default:
		var found bool
		err := types.Iterate(container, func(v interface{}) bool {
			found = (v == value)
			return !found
		})
		return found, err
	}
}

func hasKey(container, key interface{}) (bool, error) {
	switch container := container.(type) {
	case hashmap.Map:
		return hashmap.HasKey(container, key), nil
	default:
		if len := types.Len(container); len >= 0 {
			// XXX(xiaq): Not all types that implement Lener have numerical indices
			_, err := types.ConvertListIndex(key, len)
			return err == nil, nil
		}
		return false, fmt.Errorf("couldn't get key or index of type '%s'", types.Kind(container))
	}
}

func count(fm *Frame, args ...interface{}) int {
	var n int
	switch len(args) {
	case 0:
		// Count inputs.
		fm.IterateInputs(func(interface{}) {
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
	return n
}

func keys(ec *Frame, m hashmap.Map) {
	out := ec.ports[1].Chan

	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		out <- k
	}
}

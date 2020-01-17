package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
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
		"one": one,

		"has-key":   hasKey,
		"has-value": hasValue,

		"take":  take,
		"drop":  drop,
		"count": count,

		"keys": keys,
	})
}

var errKeyMustBeString = errors.New("key must be string")

func nsFn(m hashmap.Map) (Ns, error) {
	ns := make(Ns)
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			return nil, errKeyMustBeString
		}
		ns[kstring] = vars.FromInit(v)
	}
	return ns, nil
}

type rangeOpts struct{ Step float64 }

func (o *rangeOpts) SetDefaultOptions() { o.Step = 1 }

func rangeFn(fm *Frame, opts rangeOpts, args ...float64) error {
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
	for f := lower; f < upper; f += opts.Step {
		out <- vals.FromGo(f)
	}
	return nil
}

func repeat(fm *Frame, n int, v interface{}) {
	out := fm.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(fm *Frame, v interface{}) error {
	out := fm.ports[1].Chan
	return vals.Iterate(v, func(e interface{}) bool {
		out <- e
		return true
	})
}

func assoc(a, k, v interface{}) (interface{}, error) {
	return vals.Assoc(a, k, v)
}

var errCannotDissoc = errors.New("cannot dissoc")

func dissoc(a, k interface{}) (interface{}, error) {
	a2 := vals.Dissoc(a, k)
	if a2 == nil {
		return nil, errCannotDissoc
	}
	return a2, nil
}

func all(fm *Frame, inputs Inputs) {
	out := fm.ports[1].Chan
	inputs(func(v interface{}) { out <- v })
}

func one(fm *Frame, inputs Inputs) error {
	var val interface{}
	n := 0
	inputs(func(v interface{}) {
		if n == 0 {
			val = v
		}
		n++
	})
	if n == 1 {
		fm.OutputChan() <- val
		return nil
	}
	return fmt.Errorf("expect a single value, got %d", n)
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
			if vals.Equal(v, value) {
				return true, nil
			}
		}
		return false, nil
	default:
		var found bool
		err := vals.Iterate(container, func(v interface{}) bool {
			found = (v == value)
			return !found
		})
		return found, err
	}
}

func hasKey(container, key interface{}) bool {
	return vals.HasKey(container, key)
}

func count(fm *Frame, args ...interface{}) (int, error) {
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
		if len := vals.Len(v); len >= 0 {
			n = len
		} else {
			err := vals.Iterate(v, func(interface{}) bool {
				n++
				return true
			})
			if err != nil {
				return 0, fmt.Errorf("cannot get length of a %s", vals.Kind(v))
			}
		}
	default:
		return 0, errors.New("want 0 or 1 argument")
	}
	return n, nil
}

func keys(fm *Frame, v interface{}) error {
	out := fm.ports[1].Chan
	return vals.IterateKeys(v, func(k interface{}) bool {
		out <- k
		return true
	})
}

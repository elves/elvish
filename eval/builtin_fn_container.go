package eval

import (
	"errors"
	"fmt"
	"io"

	"github.com/elves/elvish/eval/vals"
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

func nsFn(m hashmap.Map) (Ns, error) {
	ns := make(Ns)
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			return nil, errKeyMustBeString
		}
		ns[kstring] = vars.NewAnyWithInit(v)
	}
	return ns, nil
}

func rangeFn(fm *Frame, rawOpts RawOptions, args ...float64) error {
	opts := struct{ Step float64 }{1}
	rawOpts.Scan(&opts)

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

func all(fm *Frame) error {
	valuesDone := make(chan struct{})
	go func() {
		for input := range fm.ports[0].Chan {
			fm.ports[1].Chan <- input
		}
		close(valuesDone)
	}()
	_, err := io.Copy(fm.ports[1].File, fm.ports[0].File)
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

func hasKey(container, key interface{}) (bool, error) {
	switch container := container.(type) {
	case hashmap.Map:
		return hashmap.HasKey(container, key), nil
	default:
		if len := vals.Len(container); len >= 0 {
			// XXX(xiaq): Not all types that implement Lener have numerical indices
			_, err := vals.ConvertListIndex(key, len)
			return err == nil, nil
		} else {
			var found bool
			err := vals.IterateKeys(container, func(k interface{}) bool {
				if key == k {
					found = true
				}
				return !found
			})
			if err == nil {
				return found, nil
			}
		}
		return false, fmt.Errorf("couldn't get key or index of type '%s'", vals.Kind(container))
	}
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

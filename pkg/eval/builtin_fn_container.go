package eval

import (
	"errors"
	"fmt"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

// Lists and maps.

func init() {
	addBuiltinFns(map[string]any{
		"ns": nsFn,

		"make-map": makeMap,

		"conj":   conj,
		"assoc":  assoc,
		"dissoc": dissoc,

		"has-key":   hasKey,
		"has-value": hasValue,

		"keys": keys,
	})
}

func nsFn(m vals.Map) (*Ns, error) {
	nb := BuildNs()
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			return nil, errs.BadValue{
				What:  `key of argument of "ns"`,
				Valid: "string", Actual: vals.Kind(k)}
		}
		nb.AddVar(kstring, vars.FromInit(v))
	}
	return nb.Ns(), nil
}

func makeMap(input Inputs) (vals.Map, error) {
	m := vals.EmptyMap
	var errMakeMap error
	input(func(v any) {
		if errMakeMap != nil {
			return
		}
		if !vals.CanIterate(v) {
			errMakeMap = errs.BadValue{
				What: "input to make-map", Valid: "iterable", Actual: vals.Kind(v)}
			return
		}
		if l := vals.Len(v); l != 2 {
			errMakeMap = errs.BadValue{
				What: "input to make-map", Valid: "iterable with 2 elements",
				Actual: fmt.Sprintf("%v with %v elements", vals.Kind(v), l)}
			return
		}
		elems, err := vals.Collect(v)
		if err != nil {
			errMakeMap = err
			return
		}
		if len(elems) != 2 {
			errMakeMap = fmt.Errorf("internal bug: collected %v values", len(elems))
			return
		}
		m = m.Assoc(elems[0], elems[1])
	})
	return m, errMakeMap
}

func conj(li vals.List, more ...any) vals.List {
	for _, val := range more {
		li = li.Conj(val)
	}
	return li
}

func assoc(a, k, v any) (any, error) {
	return vals.Assoc(a, k, v)
}

var errCannotDissoc = errors.New("cannot dissoc")

func dissoc(a, k any) (any, error) {
	a2 := vals.Dissoc(a, k)
	if a2 == nil {
		return nil, errCannotDissoc
	}
	return a2, nil
}

func hasValue(container, value any) (bool, error) {
	switch container := container.(type) {
	case vals.Map:
		for it := container.Iterator(); it.HasElem(); it.Next() {
			_, v := it.Elem()
			if vals.Equal(v, value) {
				return true, nil
			}
		}
		return false, nil
	default:
		var found bool
		err := vals.Iterate(container, func(v any) bool {
			if vals.Equal(v, value) {
				found = true
				return false
			}
			return true
		})
		return found, err
	}
}

func hasKey(container, key any) bool {
	return vals.HasKey(container, key)
}

func keys(fm *Frame, v any) error {
	out := fm.ValueOutput()
	var errPut error
	errIterate := vals.IterateKeys(v, func(k any) bool {
		errPut = out.Put(k)
		return errPut == nil
	})
	if errIterate != nil {
		return errIterate
	}
	return errPut
}

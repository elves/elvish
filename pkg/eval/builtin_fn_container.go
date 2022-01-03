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
	addBuiltinFns(map[string]interface{}{
		"ns": nsFn,

		"make-map": makeMap,

		"assoc":  assoc,
		"dissoc": dissoc,

		"has-key":   hasKey,
		"has-value": hasValue,

		"keys": keys,
	})
}

//elvdoc:fn ns
//
// ```elvish
// ns $map
// ```
//
// Constructs a namespace from `$map`, using the keys as variable names and the
// values as their values. Examples:
//
// ```elvish-transcript
// ~> var n = (ns [&name=value])
// ~> put $n[name]
// ▶ value
// ~> var n: = (ns [&name=value])
// ~> put $n:name
// ▶ value
// ```

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

//elvdoc:fn make-map
//
// ```elvish
// make-map $input?
// ```
//
// Outputs a map from the [value inputs](#value-inputs), each of which must be
// an iterable value with with two elements. The first element of each value
// is used as the key, and the second element is used as the value.
//
// If the same key appears multiple times, the last value is used.
//
// Examples:
//
// ```elvish-transcript
// ~> make-map [[k v]]
// ▶ [&k=v]
// ~> make-map [[k v1] [k v2]]
// ▶ [&k=v2]
// ~> put [k1 v1] [k2 v2] | make-map
// ▶ [&k1=v1 &k2=v2]
// ~> put aA bB | make-map
// ▶ [&a=A &b=B]
// ```

func makeMap(input Inputs) (vals.Map, error) {
	m := vals.EmptyMap
	var errMakeMap error
	input(func(v interface{}) {
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

//elvdoc:fn assoc
//
// ```elvish
// assoc $container $k $v
// ```
//
// Output a slightly modified version of `$container`, such that its value at `$k`
// is `$v`. Applies to both lists and to maps.
//
// When `$container` is a list, `$k` may be a negative index. However, slice is not
// yet supported.
//
// ```elvish-transcript
// ~> assoc [foo bar quux] 0 lorem
// ▶ [lorem bar quux]
// ~> assoc [foo bar quux] -1 ipsum
// ▶ [foo bar ipsum]
// ~> assoc [&k=v] k v2
// ▶ [&k=v2]
// ~> assoc [&k=v] k2 v2
// ▶ [&k2=v2 &k=v]
// ```
//
// Etymology: [Clojure](https://clojuredocs.org/clojure.core/assoc).
//
// @cf dissoc

func assoc(a, k, v interface{}) (interface{}, error) {
	return vals.Assoc(a, k, v)
}

var errCannotDissoc = errors.New("cannot dissoc")

//elvdoc:fn dissoc
//
// ```elvish
// dissoc $map $k
// ```
//
// Output a slightly modified version of `$map`, with the key `$k` removed. If
// `$map` does not contain `$k` as a key, the same map is returned.
//
// ```elvish-transcript
// ~> dissoc [&foo=bar &lorem=ipsum] foo
// ▶ [&lorem=ipsum]
// ~> dissoc [&foo=bar &lorem=ipsum] k
// ▶ [&lorem=ipsum &foo=bar]
// ```
//
// @cf assoc

func dissoc(a, k interface{}) (interface{}, error) {
	a2 := vals.Dissoc(a, k)
	if a2 == nil {
		return nil, errCannotDissoc
	}
	return a2, nil
}

//elvdoc:fn has-value
//
// ```elvish
// has-value $container $value
// ```
//
// Determine whether `$value` is a value in `$container`.
//
// Examples, maps:
//
// ```elvish-transcript
// ~> has-value [&k1=v1 &k2=v2] v1
// ▶ $true
// ~> has-value [&k1=v1 &k2=v2] k1
// ▶ $false
// ```
//
// Examples, lists:
//
// ```elvish-transcript
// ~> has-value [v1 v2] v1
// ▶ $true
// ~> has-value [v1 v2] k1
// ▶ $false
// ```
//
// Examples, strings:
//
// ```elvish-transcript
// ~> has-value ab b
// ▶ $true
// ~> has-value ab c
// ▶ $false
// ```

func hasValue(container, value interface{}) (bool, error) {
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
		err := vals.Iterate(container, func(v interface{}) bool {
			found = (v == value)
			return !found
		})
		return found, err
	}
}

//elvdoc:fn has-key
//
// ```elvish
// has-key $container $key
// ```
//
// Determine whether `$key` is a key in `$container`. A key could be a map key or
// an index on a list or string. This includes a range of indexes.
//
// Examples, maps:
//
// ```elvish-transcript
// ~> has-key [&k1=v1 &k2=v2] k1
// ▶ $true
// ~> has-key [&k1=v1 &k2=v2] v1
// ▶ $false
// ```
//
// Examples, lists:
//
// ```elvish-transcript
// ~> has-key [v1 v2] 0
// ▶ $true
// ~> has-key [v1 v2] 1
// ▶ $true
// ~> has-key [v1 v2] 2
// ▶ $false
// ~> has-key [v1 v2] 0:2
// ▶ $true
// ~> has-key [v1 v2] 0:3
// ▶ $false
// ```
//
// Examples, strings:
//
// ```elvish-transcript
// ~> has-key ab 0
// ▶ $true
// ~> has-key ab 1
// ▶ $true
// ~> has-key ab 2
// ▶ $false
// ~> has-key ab 0:2
// ▶ $true
// ~> has-key ab 0:3
// ▶ $false
// ```

func hasKey(container, key interface{}) bool {
	return vals.HasKey(container, key)
}

//elvdoc:fn keys
//
// ```elvish
// keys $map
// ```
//
// Put all keys of `$map` on the structured stdout.
//
// Example:
//
// ```elvish-transcript
// ~> keys [&a=foo &b=bar &c=baz]
// ▶ a
// ▶ c
// ▶ b
// ```
//
// Note that there is no guaranteed order for the keys of a map.

func keys(fm *Frame, v interface{}) error {
	out := fm.ValueOutput()
	var errPut error
	errIterate := vals.IterateKeys(v, func(k interface{}) bool {
		errPut = out.Put(k)
		return errPut == nil
	})
	if errIterate != nil {
		return errIterate
	}
	return errPut
}

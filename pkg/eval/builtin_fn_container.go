package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

// Sequence, list and maps.

// TODO(xiaq): Document "ns".

//elvdoc:fn range
//
// ```elvish
// range &step=1 $low? $high
// ```
//
// Output `$low`, `$low` + `$step`, ..., proceeding as long as smaller than
// `$high`. If not given, `$low` defaults to 0.
//
// Examples:
//
// ```elvish-transcript
// ~> range 4
// ▶ 0
// ▶ 1
// ▶ 2
// ▶ 3
// ~> range 1 6 &step=2
// ▶ 1
// ▶ 3
// ▶ 5
// ```
//
// Beware floating point oddities:
//
// ```elvish-transcript
// ~> range 0 0.8 &step=.1
// ▶ 0
// ▶ 0.1
// ▶ 0.2
// ▶ 0.30000000000000004
// ▶ 0.4
// ▶ 0.5
// ▶ 0.6
// ▶ 0.7
// ▶ 0.7999999999999999
// ```
//
// Etymology:
// [Python](https://docs.python.org/3/library/functions.html#func-range).

//elvdoc:fn repeat
//
// ```elvish
// repeat $n $value
// ```
//
// Output `$value` for `$n` times. Example:
//
// ```elvish-transcript
// ~> repeat 0 lorem
// ~> repeat 4 NAN
// ▶ NAN
// ▶ NAN
// ▶ NAN
// ▶ NAN
// ```
//
// Etymology: [Clojure](https://clojuredocs.org/clojure.core/repeat).

//elvdoc:fn explode
//
// ```elvish
// explode $iterable
// ```
//
// Put all elements of `$iterable` on the structured stdout. Like `flatten` in
// functional languages. Equivalent to `[li]{ put $@li }`.
//
// Example:
//
// ```elvish-transcript
// ~> explode [a b [x]]
// ▶ a
// ▶ b
// ▶ [x]
// ```
//
// Etymology: [PHP](http://php.net/manual/en/function.explode.php). PHP's `explode`
// is actually equivalent to Elvish's `splits`, but the author liked the name too
// much to not use it.

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

//elvdoc:fn all
//
// ```elvish $input-list?
// all
// ```
//
// Passes inputs to the output as is. Byte inputs into values, one per line.
//
// This is an identity function for commands with value outputs: `a | all` is
// equivalent to `a` if it only outputs values.
//
// This function is useful for turning inputs into arguments, like:
//
// ```elvish-transcript
// ~> put 'lorem,ipsum' | splits , (all)
// ▶ lorem
// ▶ ipsum
// ```
//
// Or capturing all inputs in a variable:
//
// ```elvish-transcript
// ~> x = [(all)]
// foo
// bar
// (Press ^D)
// ~> put $x
// ▶ [foo bar]
// ```
//
// When given a list, it outputs all elements of the list:
//
// ```elvish-transcript
// ~> all [foo bar]
// ▶ foo
// ▶ bar
// ```
//
// @cf one

//elvdoc:fn one
//
// ```elvish
// one $input-list?
// ```
//
// Passes inputs to outputs, if there is only a single one. Otherwise raises an
// exception.
//
// This function can be used in a similar way to [`all`](#all), but is a better
// choice when you expect that there is exactly one output:
//
// @cf all

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

//elvdoc:fn take
//
// ```elvish
// take $n $input-list?
// ```
//
// Retain the first `$n` input elements. If `$n` is larger than the number of input
// elements, the entire input is retained. Examples:
//
// ```elvish-transcript
// ~> take 3 [a b c d e]
// ▶ a
// ▶ b
// ▶ c
// ~> splits ' ' 'how are you?' | take 1
// ▶ how
// ~> range 2 | take 10
// ▶ 0
// ▶ 1
// ```
//
// Etymology: Haskell.

//elvdoc:fn drop
//
// ```elvish
// drop $n $input-list?
// ```
//
// Drop the first `$n` elements of the input. If `$n` is larger than the number of
// input elements, the entire input is dropped.
//
// Example:
//
// ```elvish-transcript
// ~> drop 2 [a b c d e]
// ▶ c
// ▶ d
// ▶ e
// ~> splits ' ' 'how are you?' | drop 1
// ▶ are
// ▶ 'you?'
// ~> range 2 | drop 10
// ```
//
// Etymology: Haskell.
//
// @cf take

//elvdoc:fn count
//
// ```elvish
// count $input-list?
// ```
//
// Count the number of inputs.
//
// Examples:
//
// ```elvish-transcript
// ~> count lorem # count bytes in a string
// ▶ 5
// ~> count [lorem ipsum]
// ▶ 2
// ~> range 100 | count
// ▶ 100
// ~> seq 100 | count
// ▶ 100
// ```

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

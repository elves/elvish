package eval

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/xiaq/persistent/hashmap"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

// Sequence, list and maps.

func init() {
	addBuiltinFns(map[string]interface{}{
		"ns": nsFn,

		"make-map": makeMap,

		"range":  rangeFn,
		"repeat": repeat,

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

		"order": order,
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
// ~> n = (ns [&name=value])
// ~> put $n[name]
// ▶ value
// ~> n: = (ns [&name=value])
// ~> put $n:name
// ▶ value
// ```

func nsFn(m hashmap.Map) (*Ns, error) {
	nb := make(NsBuilder)
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		kstring, ok := k.(string)
		if !ok {
			return nil, errs.BadValue{
				What:  `key of argument of "ns"`,
				Valid: "string", Actual: vals.Kind(k)}
		}
		nb[kstring] = vars.FromInit(v)
	}
	return nb.Ns(), nil
}

//elvdoc:fn make-map
//
// ```elvish
// make-map $input?
// ```
//
// Outputs a map from an input consisting of containers with two elements. The
// first element of each container is used as the key, and the second element is
// used as the value.
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

	out := fm.OutputChan()
	for f := lower; f < upper; f += opts.Step {
		out <- vals.FromGo(f)
	}
	return nil
}

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

func repeat(fm *Frame, n int, v interface{}) {
	out := fm.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
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

//elvdoc:fn all
//
// ```elvish
// all $input-list?
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
// ~> use str
// ~> put 'lorem,ipsum' | str:split , (all)
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

func all(fm *Frame, inputs Inputs) {
	out := fm.OutputChan()
	inputs(func(v interface{}) { out <- v })
}

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
// ~> use str
// ~> str:split ' ' 'how are you?' | take 1
// ▶ how
// ~> range 2 | take 10
// ▶ 0
// ▶ 1
// ```
//
// Etymology: Haskell.

func take(fm *Frame, n int, inputs Inputs) {
	out := fm.OutputChan()
	i := 0
	inputs(func(v interface{}) {
		if i < n {
			out <- v
		}
		i++
	})
}

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
// ~> use str
// ~> str:split ' ' 'how are you?' | drop 1
// ▶ are
// ▶ 'you?'
// ~> range 2 | drop 10
// ```
//
// Etymology: Haskell.
//
// @cf take

func drop(fm *Frame, n int, inputs Inputs) {
	out := fm.OutputChan()
	i := 0
	inputs(func(v interface{}) {
		if i >= n {
			out <- v
		}
		i++
	})
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

// The count implementation uses a custom varargs based implementation rather
// than the more common `Inputs` API (see pkg/eval/go_fn.go) because this
// allows the implementation to be O(1) for the common cases rather than O(n).
func count(fm *Frame, args ...interface{}) (int, error) {
	var n int
	switch nargs := len(args); nargs {
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
		// The error matches what would be returned if the `Inputs` API was
		// used. See GoFn.Call().
		return 0, errs.ArityMismatch{
			What: "arguments here", ValidLow: 0, ValidHigh: 1, Actual: nargs}
	}
	return n, nil
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
	out := fm.OutputChan()
	return vals.IterateKeys(v, func(k interface{}) bool {
		out <- k
		return true
	})
}

//elvdoc:fn order
//
// ```elvish
// order &reverse=$false $less-than=$nil $inputs?
// ```
//
// Outputs the input values sorted in ascending order. The sort is guaranteed to
// be [stable](https://en.wikipedia.org/wiki/Sorting_algorithm#Stability).
//
// The `&reverse` option, if true, reverses the order of output.
//
// The `&less-than` option, if given, establishes the ordering of the elements.
// Its value should be a function that takes two arguments and outputs a single
// boolean indicating whether the first argument is less than the second
// argument. If the function throws an exception, `order` rethrows the exception
// without outputting any value.
//
// If `&less-than` has value `$nil` (the default if not set), the following
// comparison algorithm is used:
//
// - Numbers are compared numerically. For the sake of sorting, `NaN` is treated
//   as smaller than all other numbers.
//
// - Strings are compared lexicographically by bytes, which is equivalent to
//   comparing by codepoints under UTF-8.
//
// - Lists are compared lexicographically by elements, if the elements at the
//   same positions are comparable.
//
// If the ordering between two elements are not defined by the conditions above,
// no value is outputted and an exception is thrown.
//
// Examples:
//
// ```elvish-transcript
// ~> put foo bar ipsum | order
// ▶ bar
// ▶ foo
// ▶ ipsum
// ~> order [(float64 10) (float64 1) (float64 5)]
// ▶ (float64 1)
// ▶ (float64 5)
// ▶ (float64 10)
// ~> order [[a b] [a] [b b] [a c]]
// ▶ [a]
// ▶ [a b]
// ▶ [a c]
// ▶ [b b]
// ~> order &reverse [a c b]
// ▶ c
// ▶ b
// ▶ a
// ~> order &less-than=[a b]{ eq $a x } [l x o r x e x m]
// ▶ x
// ▶ x
// ▶ x
// ▶ l
// ▶ o
// ▶ r
// ▶ e
// ▶ m
// ```
//
// Beware that strings that look like numbers are treated as strings, not
// numbers. To sort strings as numbers, use an explicit `&less-than` option:
//
// ```elvish-transcript
// ~> order [5 1 10]
// ▶ 1
// ▶ 10
// ▶ 5
// ~> order &less-than=[a b]{ < $a $b } [5 1 10]
// ▶ 1
// ▶ 5
// ▶ 10
// ```

type orderOptions struct {
	Reverse  bool
	LessThan Callable
}

func (opt *orderOptions) SetDefaultOptions() {}

// ErrUncomparable is raised by the order command when inputs contain
// uncomparable values.
var ErrUncomparable = errs.BadValue{
	What:  `inputs to "order"`,
	Valid: "comparable values", Actual: "uncomparable values"}

func order(fm *Frame, opts orderOptions, inputs Inputs) error {
	var values []interface{}
	inputs(func(v interface{}) { values = append(values, v) })

	var errSort error
	var lessFn func(i, j int) bool
	if opts.LessThan != nil {
		lessFn = func(i, j int) bool {
			if errSort != nil {
				return true
			}
			var args []interface{}
			if opts.Reverse {
				args = []interface{}{values[j], values[i]}
			} else {
				args = []interface{}{values[i], values[j]}
			}
			outputs, err := fm.CaptureOutput(func(fm *Frame) error {
				return opts.LessThan.Call(fm, args, NoOpts)
			})
			if err != nil {
				errSort = err
				return true
			}
			if len(outputs) != 1 {
				errSort = errs.BadValue{
					What:   "output of the &less-than callback",
					Valid:  "a single boolean",
					Actual: fmt.Sprintf("%d values", len(outputs))}
				return true
			}
			if b, ok := outputs[0].(bool); ok {
				return b
			}
			errSort = errs.BadValue{
				What:  "output of the &less-than callback",
				Valid: "boolean", Actual: vals.Kind(outputs[0])}
			return true
		}
	} else {
		// Use default comparison implemented by compare.
		lessFn = func(i, j int) bool {
			if errSort != nil {
				return true
			}
			o := compare(values[i], values[j])
			if o == uncomparable {
				errSort = ErrUncomparable
				return true
			}
			if opts.Reverse {
				return o == more
			}
			return o == less
		}
	}

	sort.SliceStable(values, lessFn)

	if errSort != nil {
		return errSort
	}
	for _, v := range values {
		fm.OutputChan() <- v
	}
	return nil
}

type ordering uint8

const (
	less ordering = iota
	equal
	more
	uncomparable
)

func compare(a, b interface{}) ordering {
	switch a := a.(type) {
	case float64:
		if b, ok := b.(float64); ok {
			switch {
			case math.IsNaN(a):
				if math.IsNaN(b) {
					return equal
				}
				return less
			case math.IsNaN(b):
				return more
			case a == b:
				return equal
			case a < b:
				return less
			default:
				// a > b
				return more
			}
		}
	case string:
		if b, ok := b.(string); ok {
			switch {
			case a == b:
				return equal
			case a < b:
				return less
			default:
				// a > b
				return more
			}
		}
	case vals.List:
		if b, ok := b.(vals.List); ok {
			aIt := a.Iterator()
			bIt := b.Iterator()
			for aIt.HasElem() && bIt.HasElem() {
				o := compare(aIt.Elem(), bIt.Elem())
				if o != equal {
					return o
				}
				aIt.Next()
				bIt.Next()
			}
			switch {
			case a.Len() == b.Len():
				return equal
			case a.Len() < b.Len():
				return less
			default:
				// a.Len() > b.Len()
				return more
			}
		}
	}
	return uncomparable
}

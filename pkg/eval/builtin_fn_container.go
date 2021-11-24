package eval

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/persistent/hashmap"
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
// range &step=0 $start? $end
// ```
//
// The `$start?` value defaults to zero.
//
// There are two cases depending on whether `$start` is smaller or larger than `$end`:
//
// 1) If `$start` is smaller than `$end` count up. The `&step` value must be positive and defaults
// to one. Output `$start`, `$start` + `$step`, ..., proceeding as long as the value is smaller than
// `$end` or until overflow.
//
// 1) If `$start` is larger than `$end` count down. The `&step` value must be negative and defaults
// to negative one. Output `$start`, `$start` + `$step`, ..., proceeding as long as the value is
// larger than `$end` or until overflow.
//
// Note that in both cases the `$end` value is not included in the output. The range is the
// [half-open interval](https://en.wikipedia.org/wiki/Interval_(mathematics)#Terminology)
// [`$start`, `$end`).
//
// The special `&step` value zero is replaced by -1 or +1 as appropriated for whether range is
// counting down or up respectively.
//
// This command is [exactness-preserving](#exactness-preserving). If the arguments don't make sense
// a "bad value" exception is raised.
//
// Examples:
//
// ```elvish-transcript
// ~> range 4
// ▶ (num 0)
// ▶ (num 1)
// ▶ (num 2)
// ▶ (num 3)
// ~> range 4 0
// ▶ (num 4)
// ▶ (num 3)
// ▶ (num 2)
// ▶ (num 1)
// ~> range -3 3 &step=2
// ▶ (num -3)
// ▶ (num -1)
// ▶ (num 1)
// ~> range 3 -3 &step=-2
// ▶ (num 3)
// ▶ (num 1)
// ▶ (num -1)
// ```
//
// When using floating-point numbers, beware that numerical errors can result in
// an incorrect number of outputs:
//
// ```elvish-transcript
// ~> range 0.9 &step=0.3
// ▶ (num 0.0)
// ▶ (num 0.3)
// ▶ (num 0.6)
// ▶ (num 0.8999999999999999)
// ```
//
// Avoid this problem by using exact rationals:
//
// ```elvish-transcript
// ~> range 9/10 &step=3/10
// ▶ (num 0)
// ▶ (num 3/10)
// ▶ (num 3/5)
// ```
//
// Etymology:
// [Python](https://docs.python.org/3/library/functions.html#func-range).

type rangeOpts struct{ Step vals.Num }

func (o *rangeOpts) SetDefaultOptions() { o.Step = 0 }

func rangeFn(fm *Frame, opts rangeOpts, args ...vals.Num) error {
	var rawNums []vals.Num
	switch len(args) {
	case 1:
		rawNums = []vals.Num{0, args[0], opts.Step}
	case 2:
		rawNums = []vals.Num{args[0], args[1], opts.Step}
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: 2, Actual: len(args)}
	}

	out := fm.ValueOutput()
	nums := vals.UnifyNums(rawNums, vals.Int)
	switch nums := nums.(type) {
	case []int:
		step, curValid, err := rangeInt(nums)
		if err != nil {
			return err
		}
		start, end := nums[0], nums[1]
		for prev, cur := start-step, start; curValid(prev, cur, end); cur += step {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			prev = cur
		}
	case []*big.Int:
		step, curValid, err := rangeBigInt(nums)
		if err != nil {
			return err
		}
		start, end := nums[0], nums[1]
		cur := &big.Int{}
		for cur.Set(start); curValid(cur, end); {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			next := &big.Int{}
			next.Add(cur, step)
			cur = next
		}
	case []*big.Rat:
		step, curValid, err := rangeBigRat(nums)
		if err != nil {
			return err
		}
		start, end := nums[0], nums[1]
		cur := &big.Rat{}
		for cur.Set(start); curValid(cur, end); {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			next := &big.Rat{}
			next.Add(cur, step)
			cur = next
		}
	case []float64:
		step, curValid, err := rangeFloat64(nums)
		if err != nil {
			return err
		}
		start, end := nums[0], nums[1]
		for prev, cur := start-step, start; curValid(prev, cur, end); cur += step {
			err := out.Put(vals.FromGo(cur))
			if err != nil {
				return err
			}
			prev = cur
		}
	default:
		panic("unreachable")
	}

	return nil
}

func rangeInt(nums []int) (int, func(int, int, int) bool, error) {
	start, end, step := nums[0], nums[1], nums[2]
	if step == 0 {
		if start <= end {
			step = 1
		} else {
			step = -1
		}
	}

	if step < 0 && start < end {
		return 0, nil, errs.BadValue{
			What: "step", Valid: "positive", Actual: strconv.Itoa(step)}
	}
	if step > 0 && start > end {
		return 0, nil, errs.BadValue{
			What: "step", Valid: "negative", Actual: strconv.Itoa(step)}
	}

	if step > 0 {
		curValid := func(prev, cur, end int) bool { return cur > prev && cur < end }
		return step, curValid, nil
	}
	curValid := func(prev, cur, end int) bool { return cur < prev && cur > end }
	return step, curValid, nil
}

func rangeBigInt(nums []*big.Int) (*big.Int, func(*big.Int, *big.Int) bool, error) {
	start, end, step := nums[0], nums[1], nums[2]
	if step.Sign() == 0 {
		if start.Cmp(end) <= 0 {
			step = big.NewInt(1)
		} else {
			step = big.NewInt(-1)
		}
	}

	if step.Sign() < 0 && start.Cmp(end) < 0 {
		return big.NewInt(0), nil, errs.BadValue{
			What: "step", Valid: "positive", Actual: step.String()}
	}
	if step.Sign() > 0 && start.Cmp(end) > 0 {
		return big.NewInt(0), nil, errs.BadValue{
			What: "step", Valid: "negative", Actual: step.String()}
	}

	if step.Sign() > 0 {
		curValid := func(cur, end *big.Int) bool { return cur.Cmp(end) < 0 }
		return step, curValid, nil
	}
	curValid := func(cur, end *big.Int) bool { return cur.Cmp(end) > 0 }
	return step, curValid, nil
}

func rangeBigRat(nums []*big.Rat) (*big.Rat, func(*big.Rat, *big.Rat) bool, error) {
	start, end, step := nums[0], nums[1], nums[2]
	if step.Sign() == 0 {
		if start.Cmp(end) <= 0 {
			step = big.NewRat(1, 1)
		} else {
			step = big.NewRat(-1, 1)
		}
	}

	if step.Sign() < 0 && start.Cmp(end) < 0 {
		return big.NewRat(0, 1), nil, errs.BadValue{
			What: "step", Valid: "positive", Actual: step.String()}
	}
	if step.Sign() > 0 && start.Cmp(end) > 0 {
		return big.NewRat(0, 1), nil, errs.BadValue{
			What: "step", Valid: "negative", Actual: step.String()}
	}

	if step.Sign() > 0 {
		curValid := func(cur, end *big.Rat) bool { return cur.Cmp(end) < 0 }
		return step, curValid, nil
	}
	curValid := func(cur, end *big.Rat) bool { return cur.Cmp(end) > 0 }
	return step, curValid, nil
}

func rangeFloat64(nums []float64) (float64, func(float64, float64, float64) bool, error) {
	start, end, step := nums[0], nums[1], nums[2]
	if step == 0 {
		if start <= end {
			step = 1
		} else {
			step = -1
		}
	}

	if step < 0 && start < end {
		return 0, nil, errs.BadValue{
			What: "step", Valid: "positive", Actual: vals.ToString(step)}
	}
	if step > 0 && start > end {
		return 0, nil, errs.BadValue{
			What: "step", Valid: "negative", Actual: vals.ToString(step)}
	}

	if step > 0 {
		curValid := func(prev, cur, end float64) bool { return cur > prev && cur < end }
		return step, curValid, nil
	}
	curValid := func(prev, cur, end float64) bool { return cur < prev && cur > end }
	return step, curValid, nil
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

func repeat(fm *Frame, n int, v interface{}) error {
	out := fm.ValueOutput()
	for i := 0; i < n; i++ {
		err := out.Put(v)
		if err != nil {
			return err
		}
	}
	return nil
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

func all(fm *Frame, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	inputs(func(v interface{}) {
		if errOut != nil {
			return
		}
		errOut = out.Put(v)
	})
	return errOut
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
		return fm.ValueOutput().Put(val)
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

func take(fm *Frame, n int, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	i := 0
	inputs(func(v interface{}) {
		if errOut != nil {
			return
		}
		if i < n {
			errOut = out.Put(v)
		}
		i++
	})
	return errOut
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

func drop(fm *Frame, n int, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	i := 0
	inputs(func(v interface{}) {
		if errOut != nil {
			return
		}
		if i >= n {
			errOut = out.Put(v)
		}
		i++
	})
	return errOut
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
		return 0, errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: nargs}
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
// ~> order &less-than={|a b| eq $a x } [l x o r x e x m]
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
// ~> order &less-than={|a b| < $a $b } [5 1 10]
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
	out := fm.ValueOutput()
	for _, v := range values {
		err := out.Put(v)
		if err != nil {
			return err
		}
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
	case int, *big.Int, *big.Rat, float64:
		switch b.(type) {
		case int, *big.Int, *big.Rat, float64:
			a, b := vals.UnifyNums2(a, b, 0)
			switch a := a.(type) {
			case int:
				return compareInt(a, b.(int))
			case *big.Int:
				return compareInt(a.Cmp(b.(*big.Int)), 0)
			case *big.Rat:
				return compareInt(a.Cmp(b.(*big.Rat)), 0)
			case float64:
				return compareFloat(a, b.(float64))
			default:
				panic("unreachable")
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

func compareInt(a, b int) ordering {
	if a < b {
		return less
	} else if a > b {
		return more
	}
	return equal
}

func compareFloat(a, b float64) ordering {
	// For the sake of ordering, NaN's are considered equal to each
	// other and smaller than all numbers
	switch {
	case math.IsNaN(a):
		if math.IsNaN(b) {
			return equal
		}
		return less
	case math.IsNaN(b):
		return more
	case a < b:
		return less
	case a > b:
		return more
	default: // a == b
		return equal
	}
}

package eval

import (
	"fmt"
	"sort"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Stream manipulation.

func init() {
	addBuiltinFns(map[string]interface{}{
		"all": all,
		"one": one,

		"take": take,
		"drop": drop,

		"count": count,

		"order": order,
	})
}

//elvdoc:fn all
//
// ```elvish
// all $inputs?
// ```
//
// Takes [value inputs](#value-inputs), and outputs those values unchanged.
//
// This is an [identity
// function](https://en.wikipedia.org/wiki/Identity_function) for the value
// channel; in other words, `a | all` is equivalent to just `a` if `a` only has
// value output.
//
// This command can be used inside output capture (i.e. `(all)`) to turn value
// inputs into arguments. For example:
//
// ```elvish-transcript
// ~> echo '["foo","bar"] ["lorem","ipsum"]' | from-json
// ▶ [foo bar]
// ▶ [lorem ipsum]
// ~> echo '["foo","bar"] ["lorem","ipsum"]' | from-json | put (all)[0]
// ▶ foo
// ▶ lorem
// ```
//
// The latter pipeline is equivalent to the following:
//
// ```elvish-transcript
// ~> put (echo '["foo","bar"] ["lorem","ipsum"]' | from-json)[0]
// ▶ foo
// ▶ lorem
// ```
//
// In general, when `(all)` appears in the last command of a pipeline, it is
// equivalent to just moving the previous commands of the pipeline into `()`.
// The choice is a stylistic one; the `(all)` variant is longer overall, but can
// be more readable since it allows you to avoid putting an excessively long
// pipeline inside an output capture, and keeps the data flow within the
// pipeline.
//
// Putting the value capture inside `[]` (i.e. `[(all)]`) is useful for storing
// all value inputs in a list for further processing:
//
// ```elvish-transcript
// ~> fn f { var inputs = [(all)]; put $inputs[1] }
// ~> put foo bar baz | f
// ▶ bar
// ```
//
// The `all` command can also take "inputs" from an iterable argument. This can
// be used to flatten lists or strings (although not recursively):
//
// ```elvish-transcript
// ~> all [foo [lorem ipsum]]
// ▶ foo
// ▶ [lorem ipsum]
// ~> all foo
// ▶ f
// ▶ o
// ▶ o
// ```
//
// This can be used together with `(one)` to turn a single iterable value in the
// pipeline into its elements:
//
// ```elvish-transcript
// ~> echo '["foo","bar","lorem"]' | from-json
// ▶ [foo bar lorem]
// ~> echo '["foo","bar","lorem"]' | from-json | all (one)
// ▶ foo
// ▶ bar
// ▶ lorem
// ```
//
// When given byte inputs, the `all` command currently functions like
// [`from-lines`](#from-lines), although this behavior is subject to change:
//
// ```elvish-transcript
// ~> print "foo\nbar\n" | all
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
// one $inputs?
// ```
//
// Takes exactly one [value input](#value-inputs) and outputs it. If there are
// more than one value inputs, raises an exception.
//
// This function can be used in a similar way to [`all`](#all), but is a better
// choice when you expect that there is exactly one output.
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
	return errs.ArityMismatch{What: "values", ValidLow: 1, ValidHigh: 1, Actual: n}
}

//elvdoc:fn take
//
// ```elvish
// take $n $inputs?
// ```
//
// Outputs the first `$n` [value inputs](#value-inputs). If `$n` is larger than
// the number of value inputs, outputs everything.
//
// Examples:
//
// ```elvish-transcript
// ~> range 2 | take 10
// ▶ 0
// ▶ 1
// ~> take 3 [a b c d e]
// ▶ a
// ▶ b
// ▶ c
// ~> use str
// ~> str:split ' ' 'how are you?' | take 1
// ▶ how
// ```
//
// Etymology: Haskell.
//
// @cf drop

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
// drop $n $inputs?
// ```
//
// Ignores the first `$n` [value inputs](#value-inputs) and outputs the rest.
// If `$n` is larger than the number of value inputs, outputs nothing.
//
// Example:
//
// ```elvish-transcript
// ~> range 10 | drop 8
// ▶ (num 8)
// ▶ (num 9)
// ~> range 2 | drop 10
// ~> drop 2 [a b c d e]
// ▶ c
// ▶ d
// ▶ e
// ~> use str
// ~> str:split ' ' 'how are you?' | drop 1
// ▶ are
// ▶ 'you?'
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

//elvdoc:fn order
//
// ```elvish
// order &reverse=$false $less-than=$nil $inputs?
// ```
//
// Outputs the [value inputs](#value-inputs) sorted
// in ascending order. The sorting process is guaranteed to be
// [stable](https://en.wikipedia.org/wiki/Sorting_algorithm#Stability).
//
// The `&reverse` option, if true, reverses the order of output.
//
// The `&less-than` option, if given, establishes the ordering of the elements.
// Its value should be a function that takes two arguments and outputs a single
// boolean indicating whether the first argument is less than the second
// argument. If the function throws an exception, `order` rethrows the exception
// without outputting any value.
//
// If `&less-than` has value `$nil` (the default if not set), it is equivalent
// to `{|a b| eq -1 (compare $a $b) }`.
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
//
// @cf compare

type orderOptions struct {
	Reverse  bool
	LessThan Callable
}

func (opt *orderOptions) SetDefaultOptions() {}

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
		// Use default comparison implemented by cmp.
		lessFn = func(i, j int) bool {
			if errSort != nil {
				return true
			}
			o := cmp(values[i], values[j])
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

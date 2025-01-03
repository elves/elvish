package eval

import (
	"errors"
	"fmt"
	"sort"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Stream manipulation.

func init() {
	addBuiltinFns(map[string]any{
		"all": all,
		"one": one,

		"take":    take,
		"drop":    drop,
		"compact": compact,

		"count": count,

		"order": order,

		// Iterations
		"keep-if": keepIf,
	})
}

func all(fm *Frame, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	inputs(func(v any) {
		if errOut != nil {
			return
		}
		errOut = out.Put(v)
	})
	return errOut
}

func one(fm *Frame, inputs Inputs) error {
	var val any
	n := 0
	inputs(func(v any) {
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

func take(fm *Frame, n int, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	i := 0
	inputs(func(v any) {
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

func drop(fm *Frame, n int, inputs Inputs) error {
	out := fm.ValueOutput()
	var errOut error
	i := 0
	inputs(func(v any) {
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

func compact(fm *Frame, inputs Inputs) error {
	out := fm.ValueOutput()
	first := true
	var errOut error
	var prev any

	inputs(func(v any) {
		if errOut != nil {
			return
		}
		if first || !vals.Equal(v, prev) {
			errOut = out.Put(v)
			first = false
			prev = v
		}
	})
	return errOut
}

// The count implementation uses a custom varargs based implementation rather
// than the more common `Inputs` API (see pkg/eval/go_fn.go) because this
// allows the implementation to be O(1) for the common cases rather than O(n).
func count(fm *Frame, args ...any) (int, error) {
	var n int
	switch nargs := len(args); nargs {
	case 0:
		// Count inputs.
		fm.IterateInputs(func(any) {
			n++
		})
	case 1:
		// Get length of argument.
		v := args[0]
		if len := vals.Len(v); len >= 0 {
			n = len
		} else {
			err := vals.Iterate(v, func(any) bool {
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

type orderOptions struct {
	Reverse  bool
	Key      Callable
	Total    bool
	LessThan Callable
}

func (opt *orderOptions) SetDefaultOptions() {}

// ErrBothTotalAndLessThan is returned by order when both the &total and
// &less-than options are specified.
var ErrBothTotalAndLessThan = errors.New("both &total and &less-than specified")

func order(fm *Frame, opts orderOptions, inputs Inputs) error {
	if opts.Total && opts.LessThan != nil {
		return ErrBothTotalAndLessThan
	}
	var values, keys []any
	inputs(func(v any) { values = append(values, v) })
	if opts.Key != nil {
		keys = make([]any, len(values))
		for i, value := range values {
			outputs, err := fm.CaptureOutput(func(fm *Frame) error {
				return opts.Key.Call(fm, []any{value}, NoOpts)
			})
			if err != nil {
				return err
			} else if len(outputs) != 1 {
				return errs.ArityMismatch{
					What:     "number of outputs of the &key callback",
					ValidLow: 1, ValidHigh: 1, Actual: len(outputs)}
			}
			keys[i] = outputs[0]
		}
	}

	s := &slice{fm, opts.Total, opts.LessThan, values, keys, nil}
	if opts.Reverse {
		sort.Stable(sort.Reverse(s))
	} else {
		sort.Stable(s)
	}
	if s.err != nil {
		return s.err
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

type slice struct {
	fm       *Frame
	total    bool
	lessThan Callable
	values   []any
	keys     []any // nil if no keys
	err      error
}

func (s *slice) Len() int { return len(s.values) }

func (s *slice) Less(i, j int) bool {
	if s.err != nil {
		return true
	}

	var a, b any
	if s.keys != nil {
		a, b = s.keys[i], s.keys[j]
	} else {
		a, b = s.values[i], s.values[j]
	}

	if s.lessThan == nil {
		// Use a builtin comparator depending on s.mixed.
		if s.total {
			return vals.CmpTotal(a, b) == vals.CmpLess
		}
		o := vals.Cmp(a, b)
		if o == vals.CmpUncomparable {
			s.err = ErrUncomparable
			return true
		}
		return o == vals.CmpLess
	}

	// Use the &less-than callback.
	outputs, err := s.fm.CaptureOutput(func(fm *Frame) error {
		return s.lessThan.Call(fm, []any{a, b}, NoOpts)
	})
	if err != nil {
		s.err = err
		return true
	}
	if len(outputs) != 1 {
		s.err = errs.ArityMismatch{
			What:     "number of outputs of the &less-than callback",
			ValidLow: 1, ValidHigh: 1, Actual: len(outputs)}
		return true
	}
	if b, ok := outputs[0].(bool); ok {
		return b
	}
	s.err = errs.BadValue{
		What:  "output of the &less-than callback",
		Valid: "boolean", Actual: vals.Kind(outputs[0])}
	return true
}

func (s *slice) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
	if s.keys != nil {
		s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
	}
}

func keepIf(fm *Frame, f Callable, inputs Inputs) error {
	var err error
	inputs(func(v any) {
		if err != nil {
			return
		}
		outputs, errF := fm.CaptureOutput(func(fm *Frame) error {
			return f.Call(fm, []any{v}, NoOpts)
		})
		if errF != nil {
			err = errF
		} else if len(outputs) != 1 {
			err = errs.ArityMismatch{
				What:     "number of callback outputs",
				ValidLow: 1, ValidHigh: 1, Actual: len(outputs),
			}
		} else {
			b, ok := outputs[0].(bool)
			if !ok {
				err = errs.BadValue{What: "callback output",
					Valid: "bool", Actual: vals.ReprPlain(outputs[0])}
			} else if b {
				err = fm.ValueOutput().Put(v)
			}
		}
	})
	return err
}

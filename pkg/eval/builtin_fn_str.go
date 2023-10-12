package eval

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/wcwidth"
)

// String operations.

// ErrInputOfEawkMustBeString is thrown when eawk gets a non-string input.
var ErrInputOfEawkMustBeString = errors.New("input of eawk must be string")

// TODO(xiaq): Document -override-wcswidth.

func init() {
	addBuiltinFns(map[string]any{
		"<s":  func(a, b string) bool { return a < b },
		"<=s": func(a, b string) bool { return a <= b },
		"==s": func(a, b string) bool { return a == b },
		"!=s": func(a, b string) bool { return a != b },
		">s":  func(a, b string) bool { return a > b },
		">=s": func(a, b string) bool { return a >= b },

		"to-string": toString,

		"base": base,

		"wcswidth":          wcwidth.Of,
		"-override-wcwidth": wcwidth.Override,

		"eawk": eawk,
	})
}

func toString(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.ToString(a))
		if err != nil {
			return err
		}
	}
	return nil
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(fm *Frame, b int, nums ...int) error {
	if b < 2 || b > 36 {
		return ErrBadBase
	}

	out := fm.ValueOutput()
	for _, num := range nums {
		err := out.Put(strconv.FormatInt(int64(num), b))
		if err != nil {
			return err
		}
	}
	return nil
}

type eawkOpt struct {
	Sep   string
	Posix bool
}

func (o *eawkOpt) SetDefaultOptions() {
	o.Posix = false
	o.Sep = "[ \t]+"
}

func eawk(fm *Frame, opts eawkOpt, f Callable, inputs Inputs) error {
	broken := false
	var eawkWordSep *regexp.Regexp
	var err error
	if opts.Posix {
		eawkWordSep, err = regexp.CompilePOSIX(opts.Sep)
	} else {
		eawkWordSep, err = regexp.Compile(opts.Sep)
	}
	if err != nil {
		return err
	}

	inputs(func(v any) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			broken = true
			err = ErrInputOfEawkMustBeString
			return
		}
		args := []any{line}
		for _, field := range eawkWordSep.Split(strings.Trim(line, " \t"), -1) {
			args = append(args, field)
		}

		newFm := fm.Fork("fn of eawk")
		// TODO: Close port 0 of newFm.
		ex := f.Call(newFm, args, NoOpts)
		newFm.Close()

		if ex != nil {
			switch Reason(ex) {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				broken = true
				err = ex
			}
		}
	})
	return err
}

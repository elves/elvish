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

		"eawk": Eawk,
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

// ErrInputOfEawkMustBeString is thrown when eawk gets a non-string input.
//
// TODO: Change the message to say re:awk when eawk is removed.
var ErrInputOfEawkMustBeString = errors.New("input of eawk must be string")

type eawkOpt struct {
	Sep        string
	SepPosix   bool
	SepLongest bool
}

func (o *eawkOpt) SetDefaultOptions() {
	o.Sep = "[ \t]+"
}

// Eawk implements the re:awk command and the deprecated eawk command. It is
// put in this package and exported since this package can't depend on
// src.elv.sh/pkg/mods/re.
func Eawk(fm *Frame, opts eawkOpt, f Callable, inputs Inputs) error {
	wordSep, err := makePattern(opts.Sep, opts.SepPosix, opts.SepLongest)
	if err != nil {
		return err
	}

	broken := false
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
		for _, field := range wordSep.Split(strings.Trim(line, " \t"), -1) {
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

func makePattern(p string, posix, longest bool) (*regexp.Regexp, error) {
	pattern, err := compilePattern(p, posix)
	if err != nil {
		return nil, err
	}
	if longest {
		pattern.Longest()
	}
	return pattern, nil
}

func compilePattern(pattern string, posix bool) (*regexp.Regexp, error) {
	if posix {
		return regexp.CompilePOSIX(pattern)
	}
	return regexp.Compile(pattern)
}

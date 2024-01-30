package eval

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"src.elv.sh/pkg/eval/errs"
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

func base(fm *Frame, b int, nums ...vals.Num) error {
	if b < 2 || b > 36 {
		return errs.OutOfRange{What: "base",
			ValidLow: "2", ValidHigh: "36", Actual: strconv.Itoa(b)}
	}
	// Don't output anything yet in case some arguments are invalid.
	results := make([]string, len(nums))
	for i, num := range nums {
		switch num := num.(type) {
		case int:
			results[i] = strconv.FormatInt(int64(num), b)
		case *big.Int:
			results[i] = num.Text(b)
		case float64:
			if i64 := int64(num); float64(i64) == num {
				results[i] = strconv.FormatInt(i64, b)
			} else if num == math.Trunc(num) {
				var z big.Int
				z.SetString(fmt.Sprintf("%.0f", num), 10)
				results[i] = z.Text(b)
			} else {
				return errs.BadValue{What: "number",
					Valid: "integer", Actual: vals.ReprPlain(num)}
			}
		default:
			return errs.BadValue{What: "number",
				Valid: "integer", Actual: vals.ReprPlain(num)}
		}
	}

	out := fm.ValueOutput()
	for _, s := range results {
		err := out.Put(s)
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

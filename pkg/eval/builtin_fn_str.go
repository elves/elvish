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

//elvdoc:fn &lt;s &lt;=s ==s !=s &gt;s &gt;=s {#str-cmp}
//
// ```elvish
// <s  $string... # less
// <=s $string... # less or equal
// ==s $string... # equal
// !=s $string... # not equal
// >s  $string... # greater
// >=s $string... # greater or equal
// ```
//
// String comparisons. They behave similarly to their number counterparts when
// given multiple arguments. Examples:
//
// ```elvish-transcript
// ~> >s lorem ipsum
// ▶ $true
// ~> ==s 1 1.0
// ▶ $false
// ~> >s 8 12
// ▶ $true
// ```

//elvdoc:fn wcswidth
//
// ```elvish
// wcswidth $string
// ```
//
// Output the width of `$string` when displayed on the terminal. Examples:
//
// ```elvish-transcript
// ~> wcswidth a
// ▶ 1
// ~> wcswidth lorem
// ▶ 5
// ~> wcswidth 你好，世界
// ▶ 10
// ```

// TODO(xiaq): Document -override-wcswidth.

func init() {
	addBuiltinFns(map[string]interface{}{
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

//elvdoc:fn to-string
//
// ```elvish
// to-string $value...
// ```
//
// Convert arguments to string values.
//
// ```elvish-transcript
// ~> to-string foo [a] [&k=v]
// ▶ foo
// ▶ '[a]'
// ▶ '[&k=v]'
// ```

func toString(fm *Frame, args ...interface{}) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.ToString(a))
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn base
//
// ```elvish
// base $base $number...
// ```
//
// Outputs a string for each `$number` written in `$base`. The `$base` must be
// between 2 and 36, inclusive. Examples:
//
// ```elvish-transcript
// ~> base 2 1 3 4 16 255
// ▶ 1
// ▶ 11
// ▶ 100
// ▶ 10000
// ▶ 11111111
// ~> base 16 1 3 4 16 255
// ▶ 1
// ▶ 3
// ▶ 4
// ▶ 10
// ▶ ff
// ```

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

var eawkWordSep = regexp.MustCompile("[ \t]+")

//elvdoc:fn eawk
//
// ```elvish
// eawk $f $inputs?
// ```
//
// For each [value input](#value-inputs), calls `$f` with the input followed by
// all its fields. A [`break`](./builtin.html#break) command will cause `eawk`
// to stop processing inputs. A [`continue`](./builtin.html#continue) command
// will exit $f, but is ignored by `eawk`.
//
// It should behave the same as the following functions:
//
// ```elvish
// fn eawk {|f @rest|
//   each {|line|
//     var @fields = (re:split '[ \t]+' (str:trim $line " \t"))
//     $f $line $@fields
//   } $@rest
// }
// ```
//
// This command allows you to write code very similar to `awk` scripts using
// anonymous functions. Example:
//
// ```elvish-transcript
// ~> echo " lorem ipsum\n1 2" | awk '{ print $1 }'
// lorem
// 1
// ~> echo " lorem ipsum\n1 2" | eawk {|line a b| put $a }
// ▶ lorem
// ▶ 1
// ```
//
// **Note**: Since Elvish allows variable names consisting solely of digits, you
// can also do the following:
//
// ```elvish-transcript
// ~> echo " lorem ipsum\n1 2" | eawk {|0 1 2| put $1 }
// ▶ lorem
// ▶ 1
// ```

func eawk(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			broken = true
			err = ErrInputOfEawkMustBeString
			return
		}
		args := []interface{}{line}
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

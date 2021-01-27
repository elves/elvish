package eval

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/wcwidth"
)

// String operations.

// ErrInputOfEawkMustBeString is thrown when eawk gets a non-string input.
var ErrInputOfEawkMustBeString = errors.New("input of eawk must be string")

//elvdoc:fn &lt;s &lt;=s ==s !=s &gt;s &gt;=s
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

//elvdoc:fn has-prefix
//
// ```elvish
// has-prefix $string $prefix
// ```
//
// Determine whether `$prefix` is a prefix of `$string`. Examples:
//
// ```elvish-transcript
// ~> has-prefix lorem,ipsum lor
// ▶ $true
// ~> has-prefix lorem,ipsum foo
// ▶ $false
// ```
//
// This function is deprecated; use [str:has-prefix](str.html#strhas-prefix)
// instead.

//elvdoc:fn has-suffix
//
// ```elvish
// has-suffix $string $suffix
// ```
//
// Determine whether `$suffix` is a suffix of `$string`. Examples:
//
// ```elvish-transcript
// ~> has-suffix a.html .txt
// ▶ $false
// ~> has-suffix a.html .html
// ▶ $true
// ```
//
// This function is deprecated; use [str:has-suffix](str.html#strhas-suffix)
// instead.

func init() {
	addBuiltinFns(map[string]interface{}{
		"<s":  func(a, b string) bool { return a < b },
		"<=s": func(a, b string) bool { return a <= b },
		"==s": func(a, b string) bool { return a == b },
		"!=s": func(a, b string) bool { return a != b },
		">s":  func(a, b string) bool { return a > b },
		">=s": func(a, b string) bool { return a >= b },

		"to-string": toString,

		"ord":  ord,
		"chr":  chr,
		"base": base,

		"wcswidth":          wcwidth.Of,
		"-override-wcwidth": wcwidth.Override,

		"has-prefix": strings.HasPrefix,
		"has-suffix": strings.HasSuffix,

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

func toString(fm *Frame, args ...interface{}) {
	out := fm.OutputChan()
	for _, a := range args {
		out <- vals.ToString(a)
	}
}

//elvdoc:fn ord
//
// ```elvish
// ord $string
// ```
//
// This function is deprecated; use [str:to-codepoints](str.html#strto-codepoints) instead.
//
// Output value of each codepoint in `$string`, in hexadecimal. Examples:
//
// ```elvish-transcript
// ~> ord a
// ▶ 0x61
// ~> ord 你好
// ▶ 0x4f60
// ▶ 0x597d
// ```
//
// The output format is subject to change.
//
// Etymology: [Python](https://docs.python.org/3/library/functions.html#ord).
//
// @cf chr

func ord(fm *Frame, s string) {
	out := fm.OutputChan()
	for _, r := range s {
		out <- "0x" + strconv.FormatInt(int64(r), 16)
	}
}

//elvdoc:fn chr
//
// ```elvish
// chr $number...
// ```
//
// This function is deprecated; use [str:from-codepoints](str.html#strfrom-codepoints) instead.
//
// Outputs a string consisting of the given Unicode codepoints. Example:
//
// ```elvish-transcript
// ~> chr 0x61
// ▶ a
// ~> chr 0x4f60 0x597d
// ▶ 你好
// ```
//
// Etymology: [Python](https://docs.python.org/3/library/functions.html#chr).
//
// @cf ord

func chr(nums ...int) (string, error) {
	var b bytes.Buffer
	for _, num := range nums {
		if !utf8.ValidRune(rune(num)) {
			return "", fmt.Errorf("invalid codepoint: %d", num)
		}
		b.WriteRune(rune(num))
	}
	return b.String(), nil
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

	out := fm.OutputChan()
	for _, num := range nums {
		out <- strconv.FormatInt(int64(num), b)
	}
	return nil
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

//elvdoc:fn eawk
//
// ```elvish
// eawk $f $input-list?
// ```
//
// For each input, call `$f` with the input followed by all its fields. The
// function may call `break` and `continue`.
//
// It should behave the same as the following functions:
//
// ```elvish
// fn eawk [f @rest]{
//   each [line]{
//     @fields = (re:split '[ \t]+'
//     (re:replace '^[ \t]+|[ \t]+$' '' $line))
//     $f $line $@fields
//   } $@rest
// }
// ```
//
// This command allows you to write code very similar to `awk` scripts using
// anonymous functions. Example:
//
// ```elvish-transcript
// ~> echo ' lorem ipsum
// 1 2' | awk '{ print $1 }'
// lorem
// 1
// ~> echo ' lorem ipsum
// 1 2' | eawk [line a b]{ put $a }
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

		newFm := fm.fork("fn of eawk")
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

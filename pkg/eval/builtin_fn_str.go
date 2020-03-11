package eval

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/util"
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

//elvdoc:fn ord
//
// ```elvish
// ord $string
// ```
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

//elvdoc:fn chr
//
// ```elvish
// chr $number...
// ```
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

// TODO(xiaq): Document "base".

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
// *NOTE:* Deprecated as of 0.14 and will be removed after release of 0.15.
// Please use [`str:has-prefix`](./str.html#strhas-prefix) instead.
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

//elvdoc:fn has-suffix
//
// *NOTE:* Deprecated as of 0.14 and will be removed after release of 0.15.
// Please use [`str:has-prefix`](./str.html#strhas-prefix) instead.
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

//elvdoc:fn joins
//
// ```elvish
// joins $sep $input-list?
// ```
//
// Join inputs with `$sep`. Examples:
//
// ```elvish-transcript
// ~> put lorem ipsum | joins ,
// ▶ lorem,ipsum
// ~> joins , [lorem ipsum]
// ▶ lorem,ipsum
// ```
//
// The suffix "s" means "string" and also serves to avoid colliding with the
// well-known [join](<https://en.wikipedia.org/wiki/join_(Unix)>) utility.
//
// Etymology: Various languages as `join`, in particular
// [Python](https://docs.python.org/3.6/library/stdtypes.html#str.join).
//
// @cf splits

//elvdoc:fn splits
//
// ```elvish
// splits $sep $string
// ```
//
// Split `$string` by `$sep`. If `$sep` is an empty string, split it into
// codepoints.
//
// ```elvish-transcript
// ~> splits , lorem,ipsum
// ▶ lorem
// ▶ ipsum
// ~> splits '' 你好
// ▶ 你
// ▶ 好
// ```
//
// **Note**: `splits` does not support splitting by regular expressions, `$sep` is
// always interpreted as a plain string. Use [re:split](re.html#split) if you need
// to split by regex.
//
// Etymology: Various languages as `split`, in particular
// [Python](https://docs.python.org/3.6/library/stdtypes.html#str.split).
//
// @cf joins

//elvdoc:fn replaces
//
// ```elvish
// replaces &max=-1 $old $repl $source
// ```
//
// Replace all occurrences of `$old` with `$repl` in `$source`. If `$max` is
// non-negative, it determines the max number of substitutions.
//
// **Note**: `replaces` does not support searching by regular expressions, `$old`
// is always interpreted as a plain string. Use [re:replace](re.html#replace) if
// you need to search by regex.

//elvdoc:fn eawk
//
// ```elvish
// eawk $f $input-list?
// ```
//
// For each input, call `$f` with the input followed by all its fields.
//
// It should behave the same as the following functions:
//
// ```elvish
// fn eawk [f @rest]{
// each [line]{
// @fields = (re:split '[ \t]+'
// (re:replace '^[ \t]+|[ \t]+$' '' $line))
// $f $line $@fields
// } $@rest
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

		"wcswidth":          util.Wcswidth,
		"-override-wcwidth": util.OverrideWcwidth,

		"has-prefix": strings.HasPrefix,
		"has-suffix": strings.HasSuffix,

		"joins":    joins,
		"splits":   splits,
		"replaces": replaces,

		"eawk": eawk,
	})
}

// toString converts all arguments to strings.
func toString(fm *Frame, args ...interface{}) {
	out := fm.OutputChan()
	for _, a := range args {
		out <- vals.ToString(a)
	}
}

// joins joins all input strings with a delimiter.
func joins(sep string, inputs Inputs) (string, error) {
	var buf bytes.Buffer
	var errJoin error
	first := true
	inputs(func(v interface{}) {
		if errJoin != nil {
			return
		}
		if s, ok := v.(string); ok {
			if first {
				first = false
			} else {
				buf.WriteString(sep)
			}
			buf.WriteString(s)
		} else {
			errJoin = fmt.Errorf("join wants string input, got %s", vals.Kind(v))
		}
	})
	return buf.String(), errJoin
}

type maxOpt struct{ Max int }

func (o *maxOpt) SetDefaultOptions() { o.Max = -1 }

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(fm *Frame, opts maxOpt, sep, s string) {
	out := fm.ports[1].Chan
	parts := strings.SplitN(s, sep, opts.Max)
	for _, p := range parts {
		out <- p
	}
}

func replaces(opts maxOpt, old, repl, s string) string {
	return strings.Replace(s, old, repl, opts.Max)
}

func ord(fm *Frame, s string) {
	out := fm.ports[1].Chan
	for _, r := range s {
		out <- "0x" + strconv.FormatInt(int64(r), 16)
	}
}

func chr(nums ...int) (string, error) {
	var b bytes.Buffer
	for _, num := range nums {
		if !utf8.ValidRune(rune(num)) {
			return "", fmt.Errorf("Invalid codepoint: %d", num)
		}
		b.WriteRune(rune(num))
	}
	return b.String(), nil
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(fm *Frame, b int, nums ...int) error {
	if b < 2 || b > 36 {
		return ErrBadBase
	}

	out := fm.ports[1].Chan
	for _, num := range nums {
		out <- strconv.FormatInt(int64(num), b)
	}
	return nil
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
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
		ex := newFm.Call(f, args, NoOpts)
		newFm.Close()

		if ex != nil {
			switch Cause(ex) {
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

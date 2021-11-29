// Package str exposes functionality from Go's strings package as an Elvish
// module.
package str

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

var Ns = eval.BuildNsNamed("str").
	AddGoFns(map[string]interface{}{
		"compare":      strings.Compare,
		"contains":     strings.Contains,
		"contains-any": strings.ContainsAny,
		"count":        strings.Count,
		"equal-fold":   strings.EqualFold,
		// TODO: Fields, FieldsFunc
		"from-codepoints": fromCodepoints,
		"from-utf8-bytes": fromUtf8Bytes,
		"has-prefix":      strings.HasPrefix,
		"has-suffix":      strings.HasSuffix,
		"index":           strings.Index,
		"index-any":       strings.IndexAny,
		// TODO: IndexFunc
		"join":       join,
		"last-index": strings.LastIndex,
		// TODO: LastIndexFunc, Map, Repeat
		"replace": replace,
		"split":   split,
		// TODO: SplitAfter
		"title":         strings.Title,
		"to-codepoints": toCodepoints,
		"to-lower":      strings.ToLower,
		"to-title":      strings.ToTitle,
		"to-upper":      strings.ToUpper,
		"to-utf8-bytes": toUtf8Bytes,
		// TODO: ToLowerSpecial, ToTitleSpecial, ToUpperSpecial
		"trim":       strings.Trim,
		"trim-left":  strings.TrimLeft,
		"trim-right": strings.TrimRight,
		// TODO: TrimLeft,Right}Func
		"trim-space":  strings.TrimSpace,
		"trim-prefix": strings.TrimPrefix,
		"trim-suffix": strings.TrimSuffix,
	}).Ns()

//elvdoc:fn compare
//
// ```elvish
// str:compare $a $b
// ```
//
// Compares two strings and output an integer that will be 0 if a == b,
// -1 if a < b, and +1 if a > b.
//
// ```elvish-transcript
// ~> str:compare a a
// ▶ 0
// ~> str:compare a b
// ▶ -1
// ~> str:compare b a
// ▶ 1
// ```

//elvdoc:fn contains
//
// ```elvish
// str:contains $str $substr
// ```
//
// Outputs whether `$str` contains `$substr` as a substring.
//
// ```elvish-transcript
// ~> str:contains abcd x
// ▶ $false
// ~> str:contains abcd bc
// ▶ $true
// ```

//elvdoc:fn contains-any
//
// ```elvish
// str:contains-any $str $chars
// ```
//
// Outputs whether `$str` contains any Unicode code points in `$chars`.
//
// ```elvish-transcript
// ~> str:contains-any abcd x
// ▶ $false
// ~> str:contains-any abcd xby
// ▶ $true
// ```

//elvdoc:fn count
//
// ```elvish
// str:count $str $substr
// ```
//
// Outputs the number of non-overlapping instances of `$substr` in `$s`.
// If `$substr` is an empty string, output 1 + the number of Unicode code
// points in `$s`.
//
// ```elvish-transcript
// ~> str:count abcdefabcdef bc
// ▶ 2
// ~> str:count abcdef ''
// ▶ 7
// ```

//elvdoc:fn equal-fold
//
// ```elvish
// str:equal-fold $str1 $str2
// ```
//
// Outputs if `$str1` and `$str2`, interpreted as UTF-8 strings, are equal
// under Unicode case-folding.
//
// ```elvish-transcript
// ~> str:equal-fold ABC abc
// ▶ $true
// ~> str:equal-fold abc ab
// ▶ $false
// ```

//elvdoc:fn from-codepoints
//
// ```elvish
// str:from-codepoints $number...
// ```
//
// Outputs a string consisting of the given Unicode codepoints. Example:
//
// ```elvish-transcript
// ~> str:from-codepoints 0x61
// ▶ a
// ~> str:from-codepoints 0x4f60 0x597d
// ▶ 你好
// ```
//
// @cf str:to-codepoints

func fromCodepoints(nums ...int) (string, error) {
	var b bytes.Buffer
	for _, num := range nums {
		if num < 0 || num > unicode.MaxRune {
			return "", errs.OutOfRange{
				What:     "codepoint",
				ValidLow: "0", ValidHigh: strconv.Itoa(unicode.MaxRune),
				Actual: hex(num),
			}
		}
		if !utf8.ValidRune(rune(num)) {
			return "", errs.BadValue{
				What:   "argument to str:from-codepoints",
				Valid:  "valid Unicode codepoint",
				Actual: hex(num),
			}
		}
		b.WriteRune(rune(num))
	}
	return b.String(), nil
}

func hex(i int) string {
	if i < 0 {
		return "-0x" + strconv.FormatInt(-int64(i), 16)
	}
	return "0x" + strconv.FormatInt(int64(i), 16)
}

//elvdoc:fn from-utf8-bytes
//
// ```elvish
// str:from-utf8-bytes $number...
// ```
//
// Outputs a string consisting of the given Unicode bytes. Example:
//
// ```elvish-transcript
// ~> str:from-utf8-bytes 0x61
// ▶ a
// ~> str:from-utf8-bytes 0xe4 0xbd 0xa0 0xe5 0xa5 0xbd
// ▶ 你好
// ```
//
// @cf str:to-utf8-bytes

func fromUtf8Bytes(nums ...int) (string, error) {
	var b bytes.Buffer
	for _, num := range nums {
		if num < 0 || num > 255 {
			return "", errs.OutOfRange{
				What:     "byte",
				ValidLow: "0", ValidHigh: "255",
				Actual: strconv.Itoa(num)}
		}
		b.WriteByte(byte(num))
	}
	if !utf8.Valid(b.Bytes()) {
		return "", errs.BadValue{
			What:   "arguments to str:from-utf8-bytes",
			Valid:  "valid UTF-8 sequence",
			Actual: fmt.Sprint(b.Bytes())}
	}
	return b.String(), nil
}

//elvdoc:fn has-prefix
//
// ```elvish
// str:has-prefix $str $prefix
// ```
//
// Outputs if `$str` begins with `$prefix`.
//
// ```elvish-transcript
// ~> str:has-prefix abc ab
// ▶ $true
// ~> str:has-prefix abc bc
// ▶ $false
// ```

//elvdoc:fn has-suffix
//
// ```elvish
// str:has-suffix $str $suffix
// ```
//
// Outputs if `$str` ends with `$suffix`.
//
// ```elvish-transcript
// ~> str:has-suffix abc ab
// ▶ $false
// ~> str:has-suffix abc bc
// ▶ $true
// ```

//elvdoc:fn index
//
// ```elvish
// str:index $str $substr
// ```
//
// Outputs the index of the first instance of `$substr` in `$str`, or -1
// if `$substr` is not present in `$str`.
//
// ```elvish-transcript
// ~> str:index abcd cd
// ▶ 2
// ~> str:index abcd xyz
// ▶ -1
// ```

//elvdoc:fn index-any
//
// ```elvish
// str:index-any $str $chars
// ```
//
// Outputs the index of the first instance of any Unicode code point
// from `$chars` in `$str`, or -1 if no Unicode code point from `$chars` is
// present in `$str`.
//
// ```elvish-transcript
// ~> str:index-any "chicken" "aeiouy"
// ▶ 2
// ~> str:index-any l33t aeiouy
// ▶ -1
// ```

//elvdoc:fn join
//
// ```elvish
// str:join $sep $input-list?
// ```
//
// Joins inputs with `$sep`. Examples:
//
// ```elvish-transcript
// ~> put lorem ipsum | str:join ,
// ▶ lorem,ipsum
// ~> str:join , [lorem ipsum]
// ▶ lorem,ipsum
// ~> str:join '' [lorem ipsum]
// ▶ loremipsum
// ~> str:join '...' [lorem ipsum]
// ▶ lorem...ipsum
// ```
//
// Etymology: Various languages,
// [Python](https://docs.python.org/3.6/library/stdtypes.html#str.join).
//
// @cf str:split

func join(sep string, inputs eval.Inputs) (string, error) {
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
			errJoin = errs.BadValue{
				What: "input to str:join", Valid: "string", Actual: vals.Kind(v)}
		}
	})
	return buf.String(), errJoin
}

//elvdoc:fn last-index
//
// ```elvish
// str:last-index $str $substr
// ```
//
// Outputs the index of the last instance of `$substr` in `$str`,
// or -1 if `$substr` is not present in `$str`.
//
// ```elvish-transcript
// ~> str:last-index "elven speak elvish" elv
// ▶ 12
// ~> str:last-index "elven speak elvish" romulan
// ▶ -1
// ```

//elvdoc:fn replace
//
// ```elvish
// str:replace &max=-1 $old $repl $source
// ```
//
// Replaces all occurrences of `$old` with `$repl` in `$source`. If `$max` is
// non-negative, it determines the max number of substitutions.
//
// **Note**: This command does not support searching by regular expressions, `$old`
// is always interpreted as a plain string. Use [re:replace](re.html#re:replace) if
// you need to search by regex.

type maxOpt struct{ Max int }

func (o *maxOpt) SetDefaultOptions() { o.Max = -1 }

func replace(opts maxOpt, old, repl, s string) string {
	return strings.Replace(s, old, repl, opts.Max)
}

//elvdoc:fn split
//
// ```elvish
// str:split &max=-1 $sep $string
// ```
//
// Splits `$string` by `$sep`. If `$sep` is an empty string, split it into
// codepoints.
//
// If the `&max` option is non-negative, stops after producing the maximum
// number of results.
//
// ```elvish-transcript
// ~> str:split , lorem,ipsum
// ▶ lorem
// ▶ ipsum
// ~> str:split '' 你好
// ▶ 你
// ▶ 好
// ~> str:split &max=2 ' ' 'a b c d'
// ▶ a
// ▶ 'b c d'
// ```
//
// **Note**: This command does not support splitting by regular expressions,
// `$sep` is always interpreted as a plain string. Use [re:split](re.html#re:split)
// if you need to split by regex.
//
// Etymology: Various languages, in particular
// [Python](https://docs.python.org/3.6/library/stdtypes.html#str.split).
//
// @cf str:join

func split(fm *eval.Frame, opts maxOpt, sep, s string) error {
	out := fm.ValueOutput()
	parts := strings.SplitN(s, sep, opts.Max)
	for _, p := range parts {
		err := out.Put(p)
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn title
//
// ```elvish
// str:title $str
// ```
//
// Outputs `$str` with all Unicode letters that begin words mapped to their
// Unicode title case.
//
// ```elvish-transcript
// ~> str:title "her royal highness"
// ▶ Her Royal Highness
// ```

//elvdoc:fn to-codepoints
//
// ```elvish
// str:to-codepoints $string
// ```
//
// Outputs value of each codepoint in `$string`, in hexadecimal. Examples:
//
// ```elvish-transcript
// ~> str:to-codepoints a
// ▶ 0x61
// ~> str:to-codepoints 你好
// ▶ 0x4f60
// ▶ 0x597d
// ```
//
// The output format is subject to change.
//
// @cf str:from-codepoints

func toCodepoints(fm *eval.Frame, s string) error {
	out := fm.ValueOutput()
	for _, r := range s {
		err := out.Put("0x" + strconv.FormatInt(int64(r), 16))
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn to-lower
//
// ```elvish
// str:to-lower $str
// ```
//
// Outputs `$str` with all Unicode letters mapped to their lower-case
// equivalent.
//
// ```elvish-transcript
// ~> str:to-lower 'ABC!123'
// ▶ abc!123
// ```

//elvdoc:fn to-utf8-bytes
//
// ```elvish
// str:to-utf8-bytes $string
// ```
//
// Outputs value of each byte in `$string`, in hexadecimal. Examples:
//
// ```elvish-transcript
// ~> str:to-utf8-bytes a
// ▶ 0x61
// ~> str:to-utf8-bytes 你好
// ▶ 0xe4
// ▶ 0xbd
// ▶ 0xa0
// ▶ 0xe5
// ▶ 0xa5
// ▶ 0xbd
// ```
//
// The output format is subject to change.
//
// @cf str:from-utf8-bytes

func toUtf8Bytes(fm *eval.Frame, s string) error {
	out := fm.ValueOutput()
	for _, r := range []byte(s) {
		err := out.Put("0x" + strconv.FormatInt(int64(r), 16))
		if err != nil {
			return err
		}
	}
	return nil
}

//elvdoc:fn to-title
//
// ```elvish
// str:to-title $str
// ```
//
// Outputs `$str` with all Unicode letters mapped to their Unicode title case.
//
// ```elvish-transcript
// ~> str:to-title "her royal highness"
// ▶ HER ROYAL HIGHNESS
// ~> str:to-title "хлеб"
// ▶ ХЛЕБ
// ```

//elvdoc:fn to-upper
//
// ```elvish
// str:to-upper
// ```
//
// Outputs `$str` with all Unicode letters mapped to their upper-case
// equivalent.
//
// ```elvish-transcript
// ~> str:to-upper 'abc!123'
// ▶ ABC!123
// ```

//elvdoc:fn trim
//
// ```elvish
// str:trim $str $cutset
// ```
//
// Outputs `$str` with all leading and trailing Unicode code points contained
// in `$cutset` removed.
//
// ```elvish-transcript
// ~> str:trim "¡¡¡Hello, Elven!!!" "!¡"
// ▶ 'Hello, Elven'
// ```

//elvdoc:fn trim-left
//
// ```elvish
// str:trim-left $str $cutset
// ```
//
// Outputs `$str` with all leading Unicode code points contained in `$cutset`
// removed. To remove a prefix string use [`str:trim-prefix`](#str:trim-prefix).
//
// ```elvish-transcript
// ~> str:trim-left "¡¡¡Hello, Elven!!!" "!¡"
// ▶ 'Hello, Elven!!!'
// ```

//elvdoc:fn trim-prefix
//
// ```elvish
// str:trim-prefix $str $prefix
// ```
//
// Outputs `$str` minus the leading `$prefix` string. If `$str` doesn't begin
// with `$prefix`, `$str` is output unchanged.
//
// ```elvish-transcript
// ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hello, "
// ▶ Elven!!!
// ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hola, "
// ▶ '¡¡¡Hello, Elven!!!'
// ```

//elvdoc:fn trim-right
//
// ```elvish
// str:trim-right $str $cutset
// ```
//
// Outputs `$str` with all leading Unicode code points contained in `$cutset`
// removed. To remove a suffix string use [`str:trim-suffix`](#str:trim-suffix).
//
// ```elvish-transcript
// ~> str:trim-right "¡¡¡Hello, Elven!!!" "!¡"
// ▶ '¡¡¡Hello, Elven'
// ```

//elvdoc:fn trim-space
//
// ```elvish
// str:trim-space $str
// ```
//
// Outputs `$str` with all leading and trailing white space removed as defined
// by Unicode.
//
// ```elvish-transcript
// ~> str:trim-space " \t\n Hello, Elven \n\t\r\n"
// ▶ 'Hello, Elven'
// ```

//elvdoc:fn trim-suffix
//
// ```elvish
// str:trim-suffix $str $suffix
// ```
//
// Outputs `$str` minus the trailing `$suffix` string. If `$str` doesn't end
// with `$suffix`, `$str` is output unchanged.
//
// ```elvish-transcript
// ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Elven!!!"
// ▶ ¡¡¡Hello
// ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Klingons!!!"
// ▶ '¡¡¡Hello, Elven!!!'
// ```

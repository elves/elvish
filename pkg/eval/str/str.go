// Package str exposes functionality from Go's strings package as an Elvish
// module.
package str

import (
	"strings"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
)

//elvdoc:fn compare
//
// ```elvish
// str:compare $a $b
// ```
//
// Compare returns an integer comparing two strings lexicographically. The
// result will be 0 if a==b, -1 if a < b, and +1 if a > b.
//
// ```elvish-transcript
// ~> str:compare a a
// > 0
// ~> str:compare a b
// > -1
// ~> str:compare b a
// > 1
// ```

//elvdoc:fn contains
//
// ```elvish
// str:contains $str $substr
// ```
//
// Contains returns `$true` if `$substr` is within `$str` else `$false`.
//
// ```elvish-transcript
// ~> str:contains abcd x
// > $false
// ~> str:contains abcd bc
// > $true
// ```

//elvdoc:fn contains-any
//
// ```elvish
// str:contains-any $str $chars
// ```
//
// Contains-any returns `$true` if any Unicode code points in `$chars` is
// within `$str` else `$false`.
//
// ```elvish-transcript
// ~> str:contains-any abcd x
// > $false
// ~> str:contains-any abcd xby
// > $true
// ```

//elvdoc:fn count
//
// ```elvish
// str:count $str $substr
// ```
//
// Count counts the number of non-overlapping instances of substr in s.
// If substr is an empty string, Count returns 1 + the number of Unicode code
// points in s.
//
// ```elvish-transcript
// ~> str:count abcdefabcdef bc
// > 2
// ~> str:count abcdef ''
// > 7
// ```

//elvdoc:fn equal-fold
//
// ```elvish
// str:equal-fold $str1 $str2
// ```
//
// Equal-fold reports whether `$str1` and `$str2`, interpreted as UTF-8
// strings, are equal under Unicode case-folding.
//
// ```elvish-transcript
// ~> str:equal-fold ABC abc
// > $true
// ~> str:equal-fold abc ab
// > $false
// ```

//elvdoc:fn has-prefix
//
// ```elvish
// str:has-prefix $str $prefix
// ```
//
// Has-prefix tests whether the string `$str` begins with `$prefix`
//
// ```elvish-transcript
// ~> str:has-prefix abc ab
// > $true
// ~> str:has-prefix abc bc
// > $false
// ```

//elvdoc:fn has-suffix
//
// ```elvish
// str:has-suffix $str $suffix
// ```
//
// Has-suffix tests whether the string `$str` begins with `$suffix`
//
// ```elvish-transcript
// ~> str:has-suffix abc ab
// > $false
// ~> str:has-suffix abc bc
// > $true
// ```

//elvdoc:fn index
//
// ```elvish
// str:index $str $substr
// ```
//
// Index returns the index of the first instance of `$substr` in `$str`, or -1
// if `$substr` is not present in `$str`.

//
// ```elvish-transcript
// ~> str:index abcd cd
// > 2
// ~> str:index abcd xyz
// > -1
// ```

//elvdoc:fn index-any
//
// ```elvish
// str:index-any $str $chars
// ```
//
// Index-any returns the index of the first instance of any Unicode code point
// from `$chars` in `$str`, or -1 if no Unicode code point from `$chars` is
// present in `$str`.
//
// ```elvish-transcript
// ~> str:index-any "chicken" "aeiouy"
// > 2
// ~> str:index-any l33t aeiouy
// > -1
// ```

//elvdoc:fn join
//
// ```elvish
// str:join $join_str $arg...
// ```
//
// Join concatenates its arguments separated by `$join_str`; which is
// typically a single char but can be an arbitrary string. If there are no
// arguments it reads from stdin. Lists are flattened one level. So `str:join
// - a b` produces the same output as `str:join - [a b]`. Other data types
// (such as float64) have the same representation you would get by passing the
// same arg to the `echo` builtin.
//
// ```elvish-transcript
// ~> str:join :
// > ''
// ~> str:join : x
// > x
// ~> str:join : x (float64 111.222333444555666777888999) y [a '' b] '' c [] d
// > x:111.22233344455567:y:a::b::c:d
// ~> put x (float64 111.222333444555666777888999) y [a '' b] '' c [] d | str:join :
// > x:111.22233344455567:y:a::b::c:d
// ~> echo "a\nb" | str:join -
// > a-b
// ```
//
// **Important**: Empty strings insert a join separator but empty lists do not
// since lists are flattened. Thus `str:join - a '' b` is not equivalent to
// `str:join - a [] b`.
//
// **Note**: You have to provide a join separator. If you don't it is a
// compilation error. However, you don't have to specify any arguments to be
// joined in which case it reads from stdin.

//elvdoc:fn last-index
//
// ```elvish
// str:last-index $str $substr
// ```
//
// Last-index returns the index of the last instance of `$substr` in `$str`,
// or -1 if `$substr` is not present in `$str`.
//
// ```elvish-transcript
// ~> str:last-index "elven speak elvish" elv
// > 12
// ~> str:last-index "elven speak elvish" romulan
// > -1
// ```

//elvdoc:fn title
//
// ```elvish
// str:title $str
// ```
//
// Title returns a copy of `$str` with all Unicode letters that begin words
// mapped to their Unicode title case.
//
// ```elvish-transcript
// ~> str:title "her royal highness"
// > Her Royal Highness
// ```

//elvdoc:fn to-lower
//
// ```elvish
// str:to-lower $str
// ```
//
// To-lower converts `$str` with all Unicode letters mapped to their
// lower-case equivalent.
//
// ```elvish-transcript
// ~> str:to-lower 'ABC!123'
// > abc!123
// ```

//elvdoc:fn to-title
//
// ```elvish
// str:to-title $str
// ```
//
// To-title returns a copy of `$str` with all Unicode letters mapped to
// their Unicode title case.
//
// ```elvish-transcript
// ~> str:to-title "her royal highness"
// > HER ROYAL HIGHNESS
// ~> str:to-title "хлеб"
// > ХЛЕБ
// ```

//elvdoc:fn to-upper
//
// ```elvish
// str:to-upper
// ```
//
// To-upper converts `$str` with all Unicode letters mapped to their
// upper-case equivalent.
//
// ```elvish-transcript
// ~> str:to-upper 'abc!123'
// > ABC!123
// ```

//elvdoc:fn trim
//
// ```elvish
// str:trim $str $cutset
// ```
//
// Trim returns a slice of the `$str` with all leading and trailing Unicode
// code points contained in `$cutset` removed.
//
// ```elvish-transcript
// ~> str:trim "¡¡¡Hello, Elven!!!" "!¡"
// > 'Hello, Elven'
// ```

//elvdoc:fn trim-left
//
// ```elvish
// str:trim-left $str $cutset
// ```
//
// Trim-left returns a slice of the `$str` with all leading Unicode code
// points contained in `$cutset` removed. To remove a prefix string use
// `str:trim-prefix` instead.
//
// ```elvish-transcript
// ~> str:trim-left "¡¡¡Hello, Elven!!!" "!¡"
// > 'Hello, Elven!!!'
// ```

//elvdoc:fn trim-prefix
//
// ```elvish
// str:trim-prefix $str $prefix
// ```
//
// Trim-prefix returns `$str` without the provided leading `$prefix` string.
// If `$str` doesn't begin with `$prefix`, `$str` is returned unchanged.
//
// ```elvish-transcript
// ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hello, "
// > Elven!!!
// ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hola, "
// > '¡¡¡Hello, Elven!!!'
// ```

//elvdoc:fn trim-right
//
// ```elvish
// str:trim-right $str $cutset
// ```
//
// Trim-right returns a slice of the `$str` with all leading Unicode code
// points contained in `$cutset` removed. To remove a prefix string use
// `str:trim-suffix` instead.
//
// ```elvish-transcript
// ~> str:trim-right "¡¡¡Hello, Elven!!!" "!¡"
// > '¡¡¡Hello, Elven'
// ```

//elvdoc:fn trim-space
//
// ```elvish
// str:trim-space $str
// ```
//
// Trim-space returns a copy of `$str`, with all leading and trailing white
// space removed, as defined by Unicode.
//
// ```elvish-transcript
// ~> str:trim-space " \t\n Hello, Elven \n\t\r\n"
// > 'Hello, Elven'
// ```

//elvdoc:fn trim-suffix
//
// ```elvish
// str:trim-suffix $str $suffix
// ```
//
// Trim-suffix returns `$str` without the provided trailing `$suffix` string.
// If `$str` doesn't end with `$suffix`, `$str` is returned unchanged.
//
// ```elvish-transcript
// ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Elven!!!"
// > ¡¡¡Hello
// ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Klingons!!!"
// > '¡¡¡Hello, Elven!!!'
// ```

// Wrap Go's strings.Join() so it can be used in an elvish program. We require
// the join string argument to be provided but not the strings to be joined.
// If none are provided this returns the empty string.
func strJoin(fm *eval.Frame, join_str string, args ...interface{}) string {
	var result string
	need_sep := false
	f := func(a interface{}) {
		// We could just use vals.ToString() for all arguments. However, we
		// want to flatten lists so that `str:join - a b` and `str:join - [a b]`
		// produce the same output: `a-b`.
		switch v := a.(type) {
		case string:
			// We special-case `string` since that is by far the most common
			// case, we've already determined the type, and this avoids the
			// overhead of vals.ToString().
			if need_sep {
				result += join_str
			}
			result += v
			need_sep = true
		case vals.List:
			// Note that empty lists are not the same as empty strings. The
			// latter will insert a join separator while the former will not.
			for it := v.Iterator(); it.HasElem(); it.Next() {
				if need_sep {
					result += join_str
				}
				result += vals.ToString(it.Elem())
				need_sep = true
			}
		default:
			if need_sep {
				result += join_str
			}
			result += vals.ToString(a)
			need_sep = true
		}
	}

	if len(args) > 0 {
		for _, a := range args {
			f(a)
		}
	} else {
		fm.IterateInputs(f)
	}

	return result
}

var Ns = eval.NewNs().AddGoFns("str:", fns)

var fns = map[string]interface{}{
	"compare":      strings.Compare,
	"contains":     strings.Contains,
	"contains-any": strings.ContainsAny,
	"count":        strings.Count,
	"equal-fold":   strings.EqualFold,
	// TODO: Fields, FieldsFunc
	"has-prefix": strings.HasPrefix,
	"has-suffix": strings.HasSuffix,
	"index":      strings.Index,
	"index-any":  strings.IndexAny,
	// TODO: IndexFunc
	"join":       strJoin,
	"last-index": strings.LastIndex,
	// TODO: LastIndexFunc, Map, Repeat, Replace, Split, SplitAfter
	"title":    strings.Title,
	"to-lower": strings.ToLower,
	"to-title": strings.ToTitle,
	"to-upper": strings.ToUpper,
	// TODO: ToLowerSpecial, ToTitleSpecial, ToUpperSpecial
	"trim":       strings.Trim,
	"trim-left":  strings.TrimLeft,
	"trim-right": strings.TrimRight,
	// TODO: TrimLeft,Right}Func
	"trim-space":  strings.TrimSpace,
	"trim-prefix": strings.TrimPrefix,
	"trim-suffix": strings.TrimSuffix,
}

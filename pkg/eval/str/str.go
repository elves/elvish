// Package str exposes functionality from Go's strings package as an Elvish
// module.
package str

import (
	"strings"

	"github.com/elves/elvish/pkg/eval"
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
	// TODO: IndexFunc, Join
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

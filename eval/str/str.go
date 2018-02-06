// Package str exposes functionality from Go's strings package as an Elvish
// module.
package str

import (
	"strings"

	"github.com/elves/elvish/eval"
)

var Ns = eval.NewNs().AddBuiltinFns("str:", fns)

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

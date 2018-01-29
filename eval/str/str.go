// Package str exposes functionality from Go's strings package as an Elvish
// module.
package str

import (
	"strconv"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
)

func Ns() eval.Ns {
	ns := eval.Ns{}
	eval.AddBuiltinFns(ns, fns...)
	return ns
}

var fns = []*eval.BuiltinFn{
	{"compare", wrapSSI(strings.Compare)},
	{"contains", wrapSSB(strings.Contains)},
	{"contains-any", wrapSSB(strings.ContainsAny)},
	{"count", wrapSSI(strings.Count)},
	{"equal-fold", wrapSSB(strings.EqualFold)},
	// TODO: Fields, FieldsFunc
	{"has-prefix", wrapSSB(strings.HasPrefix)},
	{"has-suffix", wrapSSB(strings.HasSuffix)},
	{"index", wrapSSI(strings.Index)},
	{"index-any", wrapSSI(strings.IndexAny)},
	// TODO: IndexFunc, Join
	{"last-index", wrapSSI(strings.LastIndex)},
	// TODO: LastIndexFunc, Map, Repeat, Replace, Split, SplitAfter
	{"title", wrapSS(strings.Title)},
	{"to-lower", wrapSS(strings.ToLower)},
	{"to-title", wrapSS(strings.ToTitle)},
	{"to-upper", wrapSS(strings.ToUpper)},
	// TODO: ToLowerSpecial, ToTitleSpecial, ToUpperSpecial
	{"trim", wrapSSS(strings.Trim)},
	{"trim-left", wrapSSS(strings.TrimLeft)},
	{"trim-right", wrapSSS(strings.TrimRight)},
	// TODO: Trim{Left,Right}Func
	{"trim-space", wrapSS(strings.TrimSpace)},
	{"trim-prefix", wrapSSS(strings.TrimPrefix)},
	{"trim-suffix", wrapSSS(strings.TrimSuffix)},
}

func wrapSS(inner func(string) string) eval.BuiltinFnImpl {
	return func(fm *eval.Frame, args []interface{}, opts map[string]interface{}) {
		var s string
		eval.ScanArgs(args, &s)
		eval.TakeNoOpt(opts)
		ret := inner(s)
		fm.OutputChan() <- ret
	}
}

func wrapSSS(inner func(a, b string) string) eval.BuiltinFnImpl {
	return func(fm *eval.Frame, args []interface{}, opts map[string]interface{}) {
		var a, b string
		eval.ScanArgs(args, &a, &b)
		eval.TakeNoOpt(opts)
		ret := inner(a, b)
		fm.OutputChan() <- ret
	}
}

func wrapSSI(inner func(a, b string) int) eval.BuiltinFnImpl {
	return func(fm *eval.Frame, args []interface{}, opts map[string]interface{}) {
		var a, b string
		eval.ScanArgs(args, &a, &b)
		eval.TakeNoOpt(opts)
		ret := inner(a, b)
		fm.OutputChan() <- strconv.Itoa(ret)
	}
}

func wrapSSB(inner func(a, b string) bool) eval.BuiltinFnImpl {
	return func(fm *eval.Frame, args []interface{}, opts map[string]interface{}) {
		var a, b string
		eval.ScanArgs(args, &a, &b)
		eval.TakeNoOpt(opts)
		ret := inner(a, b)
		fm.OutputChan() <- types.Bool(ret)
	}
}

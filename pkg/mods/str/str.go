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
	AddGoFns(map[string]any{
		"compare":      strings.Compare,
		"contains":     strings.Contains,
		"contains-any": strings.ContainsAny,
		"count":        strings.Count,
		"equal-fold":   strings.EqualFold,
		// TODO: FieldsFunc
		"fields":          strings.Fields,
		"from-codepoints": fromCodepoints,
		"from-utf8-bytes": fromUtf8Bytes,
		"has-prefix":      strings.HasPrefix,
		"has-suffix":      strings.HasSuffix,
		"index":           strings.Index,
		"index-any":       strings.IndexAny,
		// TODO: IndexFunc
		"join":       join,
		"last-index": strings.LastIndex,
		// TODO: LastIndexFunc, Map
		"repeat":  repeat,
		"replace": replace,
		"split":   split,
		// TODO: SplitAfter
		//lint:ignore SA1019 Elvish builtins need to be formally deprecated
		// before removal
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

func join(sep string, inputs eval.Inputs) (string, error) {
	var buf bytes.Buffer
	var errJoin error
	first := true
	inputs(func(v any) {
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

func repeat(s string, n int) (string, error) {
	if n < 0 {
		return "", errs.BadValue{What: "n", Valid: "non-negative number", Actual: vals.ToString(n)}
	}
	if len(s)*n < 0 {
		return "", errs.BadValue{What: "n", Valid: "small enough not to overflow result", Actual: vals.ToString(n)}
	}
	return strings.Repeat(s, n), nil
}

type maxOpt struct{ Max int }

func (o *maxOpt) SetDefaultOptions() { o.Max = -1 }

func replace(opts maxOpt, old, repl, s string) string {
	return strings.Replace(s, old, repl, opts.Max)
}

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

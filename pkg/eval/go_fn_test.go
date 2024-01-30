package eval_test

import (
	"errors"
	"fmt"
	"math/big"

	. "src.elv.sh/pkg/eval"
)

type someOptions struct {
	Foo string
	Bar string
}

func (o *someOptions) SetDefaultOptions() { o.Bar = "default" }

type namedSlice []string

var goFnsMod = BuildNs().AddGoFns(map[string]any{
	"nullary": func() {},
	"takes-two-strings": func(fm *Frame, a, b string) {
		fmt.Fprintf(fm.ByteOutput(), "a = %q, b = %q\n", a, b)
	},
	"takes-variadic-strings": func(fm *Frame, args ...string) {
		fmt.Fprintf(fm.ByteOutput(), "args = %q\n", args)
	},
	"takes-string-and-variadic-strings": func(fm *Frame, first string, more ...string) {
		fmt.Fprintf(fm.ByteOutput(), "first = %q, more = %q\n", first, more)
	},
	"takes-int-float64": func(fm *Frame, i int, f float64) {
		fmt.Fprintf(fm.ByteOutput(), "i = %v, f = %v\n", i, f)
	},
	"takes-input": func(fm *Frame, i Inputs) {
		i(func(x any) {
			fmt.Fprintf(fm.ByteOutput(), "input: %v\n", x)
		})
	},
	"takes-options": func(fm *Frame, opts someOptions) {
		fmt.Fprintf(fm.ByteOutput(), "opts = %#v\n", opts)
	},
	"takes-raw-options": func(fm *Frame, opts RawOptions) {
		fmt.Fprintf(fm.ByteOutput(), "opts = %#v\n", opts)
	},
	"returns-string":        func() string { return "a string" },
	"returns-int":           func() int { return 233 },
	"returns-small-big-int": func() *big.Int { return big.NewInt(233) },

	"returns-slice": func() []string { return []string{"foo", "bar"} },
	"returns-array": func() [2]string { return [2]string{"foo", "bar"} },
	"returns-named-slice-type": func() namedSlice {
		return namedSlice{"foo", "bar"}
	},

	"returns-non-nil-error": func() error { return errors.New("bad") },
	"returns-nil-error":     func() error { return nil },
})

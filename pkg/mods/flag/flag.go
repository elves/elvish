package flag

import (
	"errors"
	"flag"
	"io"
	"math/big"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/getopt"
)

// Ns is the namespace for the flag: module.
var Ns = eval.BuildNsNamed("flag").
	AddGoFns(map[string]interface{}{
		"call":         call,
		"parse":        parse,
		"parse-getopt": parseGetopt,
	}).Ns()

//elvdoc:fn call
//
// ```elvish
// flag:call $fn $args
// ```
//
// Parses flags from `$args` according to the signature of the
// `$fn`, using the [Go convention](#go-convention), and calls `$fn`.
//
// The `$fn` must be a user-defined function (i.e. not a builtin
// function or external command). Each option corresponds to a flag; see
// [`flag:parse`](#flag:parse) for how the default value affects the behavior of
// flags. After parsing, the non-flag arguments are used as function arguments.
//
// Example:
//
// ```elvish-transcript
// ~> use flag
// ~> fn f {|&verbose=$false &port=(num 8000) name| put $verbose $port $name }
// ~> flag:call $f [-verbose -port 80 a.c]
// ▶ $true
// ▶ (num 80)
// ▶ a.c
// ```
//
// @cf flag:parse

func call(fm *eval.Frame, fn *eval.Closure, argsVal vals.List) error {
	var args []string
	err := vals.ScanListToGo(argsVal, &args)
	if err != nil {
		return err
	}
	fs := newFlagSet("")
	for i, name := range fn.OptNames {
		value := fn.OptDefaults[i]
		addFlag(fs, name, value, "")
	}
	err = fs.Parse(args)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	fs.VisitAll(func(f *flag.Flag) {
		m[f.Name] = f.Value.(flag.Getter).Get()
	})
	return fn.Call(fm.Fork("parse:call"), callArgs(fs.Args()), m)
}

func callArgs(ss []string) []interface{} {
	vs := make([]interface{}, len(ss))
	for i, s := range ss {
		vs[i] = s
	}
	return vs
}

//elvdoc:fn parse
//
// ```elvish
// flag:parse $args $specs
// ```
//
// Parses flags from `$args` according to the `$specs`, using the [Go
// convention](#go-convention).
//
// The `$args` must be a list of strings containing the command-line arguments
// to parse.
//
// The `$specs` must be a list of flag specs:
//
// ```elvish
// [
//   [flag default-value 'description of the flag']
//   ...
// ]
// ```
//
// Each flag spec consists of the name of the flag (without the leading `-`),
// its default value, and a description. The default value influences the how
// the flag gets converted from string:
//
// -   If it is boolean, the flag is a boolean flag (see [Go
//     convention](#go-convention) for implications). Flag values `0`, `f`, `F`,
//     `false`, `False` and `FALSE` are converted to `$false`, and `1`, `t`,
//     `T`, `true`, `True` and `TRUE` to `$true`. Other values are invalid.
//
// -   If it is a string, no conversion is done.
//
// -   If it is a [typed number](language.html#number), the flag value is
//     converted using [`num`](builtin.html#num).
//
// -   If it is a list, the flag value is split at `,` (equivalent to `{|s| put
//     [(str:split , $s)] }`).
//
// -   If it is none of the above, an exception is thrown.
//
// On success, this command outputs two values: a map containing the value of
// flags defined in `$specs` (whether they appear in `$args` or not), and a list
// containing non-flag arguments.
//
// Example:
//
// ```elvish-transcript
// ~> flag:parse [-v -times 10 foo] [
//      [v $false 'Verbose']
//      [times (num 1) 'How many times']
//    ]
// ▶ [&v=$true &times=(num 10)]
// ▶ [foo]
// ~> flag:parse [] [
//      [v $false 'Verbose']
//      [times (num 1) 'How many times']
//    ]
// ▶ [&v=$false &times=(num 1)]
// ▶ []
// ```
//
// @cf flag:call flag:parse-getopt

func parse(argsVal vals.List, specsVal vals.List) (vals.Map, vals.List, error) {
	var args []string
	err := vals.ScanListToGo(argsVal, &args)
	if err != nil {
		return nil, nil, err
	}
	var specs []vals.List
	err = vals.ScanListToGo(specsVal, &specs)
	if err != nil {
		return nil, nil, err
	}

	fs := newFlagSet("")
	for _, spec := range specs {
		var (
			name        string
			value       interface{}
			description string
		)
		vals.ScanListElementsToGo(spec, &name, &value, &description)
		err := addFlag(fs, name, value, description)
		if err != nil {
			return nil, nil, err
		}
	}
	err = fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}
	m := vals.EmptyMap
	fs.VisitAll(func(f *flag.Flag) {
		m = m.Assoc(f.Name, f.Value.(flag.Getter).Get())
	})
	return m, vals.MakeListFromStrings(fs.Args()...), nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func addFlag(fs *flag.FlagSet, name string, value interface{}, description string) error {
	switch value := value.(type) {
	case bool:
		fs.Bool(name, value, description)
	case string:
		fs.String(name, value, description)
	case int, *big.Int, *big.Rat, float64:
		fs.Var(&numFlag{value}, name, description)
	case vals.List:
		fs.Var(&listFlag{value}, name, description)
	default:
		return errs.BadValue{What: "flag default value",
			Valid:  "boolean, number, string or list",
			Actual: vals.ReprPlain(value)}
	}
	return nil
}

type numFlag struct{ value vals.Num }

func (nf *numFlag) String() string     { return vals.ToString(nf.value) }
func (nf *numFlag) Get() interface{}   { return nf.value }
func (nf *numFlag) Set(s string) error { return vals.ScanToGo(s, &nf.value) }

type listFlag struct{ value vals.List }

func (lf *listFlag) String() string   { return vals.ToString(lf.value) }
func (lf *listFlag) Get() interface{} { return lf.value }

func (lf *listFlag) Set(s string) error {
	lf.value = vals.MakeListFromStrings(strings.Split(s, ",")...)
	return nil
}

//elvdoc:fn parse-getopt
//
// ```elvish
// flag:parse-getopt $args $specs ^
//   &stop-after-double-dash=$true &stop-before-non-flag=$false &long-only=$false
// ```
//
// Parses flags from `$args` according to the `$specs`, using the [getopt
// convention](#getopt-convention) (see there for the semantics of the options),
// and outputs the result.
//
// The `$args` must be a list of strings containing the command-line arguments
// to parse.
//
// The `$specs` must be a list of flag specs:
//
// ```elvish
// [
//   [&short=f &long=flag &arg-optional=$false &arg-required=$false]
//   ...
// ]
// ```
//
// Each flag spec can contain the following:
//
// -   The short and long form of the flag, without the leading `-` or `--`. The
//     short form, if non-empty, must be one character. At least one of `&short`
//     and `&long` must be non-empty.
//
// -   Whether the flag takes an optional argument or a required argument. At
//     most one of `&arg-optional` and `&arg-required` may be true.
//
// It is not an error for a flag spec to contain more keys.
//
// On success, this command outputs two values: a list describing all flags
// parsed from `$args`, and a list containing non-flag arguments. The former
// list looks like:
//
// ```elvish
// [
//   [&spec=... &arg=value &long=$false]
//   ...
// ]
// ```
//
// Each entry contains the original spec for the flag, its argument, and whether
// the flag appeared in its long form.
//
// Example (some output reformatted for readability):
//
// ```elvish-transcript
// ~> var specs = [
//      [&short=v &long=verbose]
//      [&short=p &long=port &arg-required]
//    ]
// ~> flag:parse-getopt [-v -p 80 foo] $specs
// ▶ [[&spec=[&short=v &long=verbose] &long=$false &arg='']
//    [&spec=[&arg-required=$true &short=p &long=port] &long=$false &arg=80]]
// ▶ [foo]
// ~> flag:parse-getopt [--verbose] $specs
// ▶ [[&spec=[&short=v &long=verbose] &long=$true &arg='']]
// ▶ []
// ~> flag:parse-getopt [-v] [[&short=v &extra-info=foo]] # extra key in spec
// ▶ [[&spec=[&extra-info=foo &short=v] &long=$false &arg='']]
// ▶ []
// ```
//
// @cf flag:parse edit:complete-getopt

type specStruct struct {
	Short       rune
	Long        string
	ArgRequired bool
	ArgOptional bool
}

var (
	errShortLong              = errors.New("at least one of &short and &long must be non-empty")
	errArgRequiredArgOptional = errors.New("at most one of &arg-required and &arg-optional may be true")
)

func (s *specStruct) OptionSpec() (*getopt.OptionSpec, error) {
	if s.Short == 0 && s.Long == "" {
		return nil, errShortLong
	}
	arity := getopt.NoArgument
	switch {
	case s.ArgRequired && s.ArgOptional:
		return nil, errArgRequiredArgOptional
	case s.ArgRequired:
		arity = getopt.RequiredArgument
	case s.ArgOptional:
		arity = getopt.OptionalArgument
	}
	return &getopt.OptionSpec{Short: s.Short, Long: s.Long, Arity: arity}, nil
}

type parseGetoptOptions struct {
	StopAfterDoubleDash bool
	StopBeforeNonFlag   bool
	LongOnly            bool
}

func (o *parseGetoptOptions) SetDefaultOptions() { o.StopAfterDoubleDash = true }

func (o *parseGetoptOptions) Config() getopt.Config {
	c := getopt.Config(0)
	if o.StopAfterDoubleDash {
		c |= getopt.StopAfterDoubleDash
	}
	if o.StopBeforeNonFlag {
		c |= getopt.StopBeforeFirstNonOption
	}
	if o.LongOnly {
		c |= getopt.LongOnly
	}
	return c
}

func parseGetopt(opts parseGetoptOptions, argsVal vals.List, specsVal vals.List) (vals.List, vals.List, error) {
	var args []string
	err := vals.ScanListToGo(argsVal, &args)
	if err != nil {
		return nil, nil, err
	}
	var specMaps []vals.Map
	err = vals.ScanListToGo(specsVal, &specMaps)
	if err != nil {
		return nil, nil, err
	}

	specs := make([]*getopt.OptionSpec, len(specMaps))
	originalSpecMap := make(map[*getopt.OptionSpec]vals.Map)
	for i, specMap := range specMaps {
		var s specStruct
		vals.ScanMapToGo(specMap, &s)
		spec, err := s.OptionSpec()
		if err != nil {
			return nil, nil, err
		}
		specs[i] = spec
		originalSpecMap[spec] = specMap
	}
	flags, nonFlagArgs, err := getopt.Parse(args, specs, opts.Config())
	if err != nil {
		return nil, nil, err
	}

	flagsList := vals.EmptyList
	for _, flag := range flags {
		flagsList = flagsList.Conj(
			vals.MakeMap(
				"spec", originalSpecMap[flag.Spec],
				"arg", flag.Argument,
				"long", flag.Long))
	}

	return flagsList, vals.MakeListFromStrings(nonFlagArgs...), nil
}

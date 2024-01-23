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
	AddGoFns(map[string]any{
		"call":         call,
		"parse":        parse,
		"parse-getopt": parseGetopt,
	}).Ns()

func call(fm *eval.Frame, fn *eval.Closure, argsVal vals.List) error {
	var args []string
	err := vals.ScanListToGo(argsVal, &args)
	if err != nil {
		return err
	}
	fs := newFlagSet("")
	for i, name := range fn.OptNames {
		value := fn.OptDefaults[i]
		err := addFlag(fs, name, value, "")
		if err != nil {
			return err
		}
	}
	err = fs.Parse(args)
	if err != nil {
		return err
	}
	m := make(map[string]any)
	fs.VisitAll(func(f *flag.Flag) {
		m[f.Name] = f.Value.(flag.Getter).Get()
	})
	return fn.Call(fm.Fork("parse:call"), callArgs(fs.Args()), m)
}

func callArgs(ss []string) []any {
	vs := make([]any, len(ss))
	for i, s := range ss {
		vs[i] = s
	}
	return vs
}

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
			value       any
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
	return m, vals.MakeListSlice(fs.Args()), nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func addFlag(fs *flag.FlagSet, name string, value any, description string) error {
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
func (nf *numFlag) Get() any           { return nf.value }
func (nf *numFlag) Set(s string) error { return vals.ScanToGo(s, &nf.value) }

type listFlag struct{ value vals.List }

func (lf *listFlag) String() string { return vals.ToString(lf.value) }
func (lf *listFlag) Get() any       { return lf.value }

func (lf *listFlag) Set(s string) error {
	lf.value = vals.MakeListSlice(strings.Split(s, ","))
	return nil
}

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

	return flagsList, vals.MakeListSlice(nonFlagArgs), nil
}

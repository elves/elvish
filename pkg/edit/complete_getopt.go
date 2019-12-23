package edit

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/getopt"
	"github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/hashmap"
)

//elvdoc:fn complete-getopt
//
// ```elvish
// edit:complete-getopt $args $opt-specs $arg-handlers
// ```
// Produces completions according to a specification of accepted command-line
// options (both short and long options are handled), positional handler
// functions for each command position, and the current arguments in the command
// line. The arguments are as follows:
//
// * `$args` is an array containing the current arguments in the command line
//   (without the command itself). These are the arguments as passed to the
//   [Argument Completer](#argument-completer) function.
//
// * `$opt-specs` is an array of maps, each one containing the definition of
//   one possible command-line option. Matching options will be provided as
//   completions when the last element of `$args` starts with a dash, but not
//   otherwise. Each map can contain the following keys (at least one of `short`
//   or `long` needs to be specified):
//
//   - `short` contains the one-letter short option, if any, without the dash.
//
//   - `long` contains the long option name, if any, without the initial two
//     dashes.
//
//   - `arg-optional`, if set to `$true`, specifies that the option receives an
//     optional argument.
//
//   - `arg-required`, if set to `$true`, specifies that the option receives a
//     mandatory argument. Only one of `arg-optional` or `arg-required` can be
//     set to `$true`.
//
//   - `desc` can be set to a human-readable description of the option which
//     will be displayed in the completion menu.
//
//   - `completer` can be set to a function to generate possible completions for
//     the option argument. The function receives as argument the element at
//     that position and return zero or more candidates.
//
// * `$arg-handlers` is an array of functions, each one returning the possible
//   completions for that position in the arguments. Each function receives
//   as argument the last element of `$args`, and should return zero or more
//   possible values for the completions at that point. The returned values can
//   be plain strings or the output of `edit:complex-candidate`. If the last
//   element of the list is the string `...`, then the last handler is reused
//   for all following arguments.
//
// Example:
//
// ```elvish-transcript
// ~> fn complete [@args]{
//      opt-specs = [ [&short=a &long=all &desc="Show all"]
//                    [&short=n &desc="Set name" &arg-required=$true
//                     &completer= [_]{ put name1 name2 }] ]
//      arg-handlers = [ [_]{ put first1 first2 }
//                       [_]{ put second1 second2 } ... ]
//      edit:complete-getopt $args $opt-specs $arg-handlers
//    }
// ~> complete ''
// ▶ first1
// ▶ first2
// ~> complete '-'
// ▶ (edit:complex-candidate -a &display-suffix=' (Show all)')
// ▶ (edit:complex-candidate --all &display-suffix=' (Show all)')
// ▶ (edit:complex-candidate -n &display-suffix=' (Set name)')
// ~> complete -n ''
// ▶ name1
// ▶ name2
// ~> complete -a ''
// ▶ first1
// ▶ first2
// ~> complete arg1 ''
// ▶ second1
// ▶ second2
// ~> complete arg1 arg2 ''
// ▶ second1
// ▶ second2
// ```

func completeGetopt(fm *eval.Frame, vArgs, vOpts, vArgHandlers interface{}) error {
	args, err := parseGetoptArgs(vArgs)
	if err != nil {
		return err
	}
	opts, err := parseGetoptOptSpecs(vOpts)
	if err != nil {
		return err
	}
	argHandlers, variadic, err := parseGetoptArgHandlers(vArgHandlers)
	if err != nil {
		return err
	}

	// TODO(xiaq): Make the Config field configurable
	g := getopt.Getopt{Options: opts.opts, Config: getopt.GNUGetoptLong}
	_, parsedArgs, ctx := g.Parse(args)

	out := fm.OutputChan()
	putShortOpt := func(opt *getopt.Option) {
		c := complexItem{Stem: "-" + string(opt.Short)}
		if d, ok := opts.desc[opt]; ok {
			if e, ok := opts.argDesc[opt]; ok {
				c.DisplaySuffix = " " + e + " (" + d + ")"
			} else {
				c.DisplaySuffix = " (" + d + ")"
			}
		}
		out <- c
	}
	putLongOpt := func(opt *getopt.Option) {
		c := complexItem{Stem: "--" + opt.Long}
		if d, ok := opts.desc[opt]; ok {
			if e, ok := opts.argDesc[opt]; ok {
				c.DisplaySuffix = " " + e + " (" + d + ")"
			} else {
				c.DisplaySuffix = " (" + d + ")"
			}
		}
		out <- c
	}
	call := func(fn eval.Callable, args ...interface{}) {
		fm.Call(fn, args, eval.NoOpts)
	}

	switch ctx.Type {
	case getopt.NewOptionOrArgument, getopt.Argument:
		// Find argument handler.
		var argHandler eval.Callable
		if len(parsedArgs) < len(argHandlers) {
			argHandler = argHandlers[len(parsedArgs)]
		} else if variadic {
			argHandler = argHandlers[len(argHandlers)-1]
		}
		if argHandler != nil {
			call(argHandler, ctx.Text)
		} else {
			// TODO(xiaq): Notify that there is no suitable argument completer.
		}
	case getopt.NewOption:
		for _, opt := range opts.opts {
			if opt.Short != 0 {
				putShortOpt(opt)
			}
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.NewLongOption:
		for _, opt := range opts.opts {
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.LongOption:
		for _, opt := range opts.opts {
			if strings.HasPrefix(opt.Long, ctx.Text) {
				putLongOpt(opt)
			}
		}
	case getopt.ChainShortOption:
		for _, opt := range opts.opts {
			if opt.Short != 0 {
				// XXX loses chained options
				putShortOpt(opt)
			}
		}
	case getopt.OptionArgument:
		gen := opts.argGenerator[ctx.Option.Option]
		if gen != nil {
			call(gen, ctx.Option.Argument)
		}
	}
	return nil
}

// TODO(xiaq): Simplify most of the parsing below with reflection.

func parseGetoptArgs(v interface{}) ([]string, error) {
	var args []string
	var err error
	errIterate := vals.Iterate(v, func(v interface{}) bool {
		arg, ok := v.(string)
		if !ok {
			err = fmt.Errorf("arg should be string, got %s", vals.Kind(v))
			return false
		}
		args = append(args, arg)
		return true
	})
	if errIterate != nil {
		err = errIterate
	}
	return args, err
}

type parsedOptSpecs struct {
	opts         []*getopt.Option
	desc         map[*getopt.Option]string
	argDesc      map[*getopt.Option]string
	argGenerator map[*getopt.Option]eval.Callable
}

func parseGetoptOptSpecs(v interface{}) (parsedOptSpecs, error) {
	result := parsedOptSpecs{
		nil, map[*getopt.Option]string{},
		map[*getopt.Option]string{}, map[*getopt.Option]eval.Callable{}}

	var err error
	errIterate := vals.Iterate(v, func(v interface{}) bool {
		m, ok := v.(hashmap.Map)
		if !ok {
			err = fmt.Errorf("opt should be map, got %s", vals.Kind(v))
			return false
		}

		opt := &getopt.Option{}

		getStringField := func(k string) (string, bool, error) {
			v, ok := m.Index(k)
			if !ok {
				return "", false, nil
			}
			if vs, ok := v.(string); ok {
				return vs, true, nil
			}
			return "", false,
				fmt.Errorf("%s should be string, got %s", k, vals.Kind(v))
		}
		getCallableField := func(k string) (eval.Callable, bool, error) {
			v, ok := m.Index(k)
			if !ok {
				return nil, false, nil
			}
			if vb, ok := v.(eval.Callable); ok {
				return vb, true, nil
			}
			return nil, false,
				fmt.Errorf("%s should be fn, got %s", k, vals.Kind(v))
		}
		getBoolField := func(k string) (bool, bool, error) {
			v, ok := m.Index(k)
			if !ok {
				return false, false, nil
			}
			if vb, ok := v.(bool); ok {
				return vb, true, nil
			}
			return false, false,
				fmt.Errorf("%s should be bool, got %s", k, vals.Kind(v))
		}

		if s, ok, errGet := getStringField("short"); ok {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || size != len(s) {
				err = fmt.Errorf(
					"short option should be exactly one rune, got %v",
					parse.Quote(s))
				return false
			}
			opt.Short = r
		} else if errGet != nil {
			err = errGet
			return false
		}
		if s, ok, errGet := getStringField("long"); ok {
			opt.Long = s
		} else if errGet != nil {
			err = errGet
			return false
		}
		if opt.Short == 0 && opt.Long == "" {
			err = errors.New(
				"opt should have at least one of short and long forms")
			return false
		}

		argRequired, _, errGet := getBoolField("arg-required")
		if errGet != nil {
			err = errGet
			return false
		}
		argOptional, _, errGet := getBoolField("arg-optional")
		if errGet != nil {
			err = errGet
			return false
		}
		switch {
		case argRequired && argOptional:
			err = errors.New(
				"opt cannot have both arg-required and arg-optional")
			return false
		case argRequired:
			opt.HasArg = getopt.RequiredArgument
		case argOptional:
			opt.HasArg = getopt.OptionalArgument
		}

		if s, ok, errGet := getStringField("desc"); ok {
			result.desc[opt] = s
		} else if errGet != nil {
			err = errGet
			return false
		}
		if s, ok, errGet := getStringField("arg-desc"); ok {
			result.argDesc[opt] = s
		} else if errGet != nil {
			err = errGet
			return false
		}
		if f, ok, errGet := getCallableField("completer"); ok {
			result.argGenerator[opt] = f
		} else if errGet != nil {
			err = errGet
			return false
		}

		result.opts = append(result.opts, opt)
		return true
	})
	if errIterate != nil {
		err = errIterate
	}
	return result, err
}

func parseGetoptArgHandlers(v interface{}) ([]eval.Callable, bool, error) {
	var argHandlers []eval.Callable
	var variadic bool
	var err error
	errIterate := vals.Iterate(v, func(v interface{}) bool {
		sv, ok := v.(string)
		if ok {
			if sv == "..." {
				variadic = true
				return true
			}
			err = fmt.Errorf(
				"string except for ... not allowed as argument handler, got %s",
				parse.Quote(sv))
			return false
		}
		argHandler, ok := v.(eval.Callable)
		if !ok {
			err = fmt.Errorf(
				"argument handler should be fn, got %s", vals.Kind(v))
		}
		argHandlers = append(argHandlers, argHandler)
		return true
	})
	if errIterate != nil {
		err = errIterate
	}
	return argHandlers, variadic, err
}

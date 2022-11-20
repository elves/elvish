package edit

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/getopt"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

func completeGetopt(fm *eval.Frame, vArgs, vOpts, vArgHandlers any) error {
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

	// TODO: Make the Config field configurable
	_, parsedArgs, ctx := getopt.Complete(args, opts.opts, getopt.GNU)

	out := fm.ValueOutput()
	putShortOpt := func(opt *getopt.OptionSpec) error {
		c := complexItem{Stem: "-" + string(opt.Short)}
		if d, ok := opts.desc[opt]; ok {
			if e, ok := opts.argDesc[opt]; ok {
				c.Display = ui.T(c.Stem + " " + e + " (" + d + ")")
			} else {
				c.Display = ui.T(c.Stem + " (" + d + ")")
			}
		}
		return out.Put(c)
	}
	putLongOpt := func(opt *getopt.OptionSpec) error {
		c := complexItem{Stem: "--" + opt.Long}
		if d, ok := opts.desc[opt]; ok {
			if e, ok := opts.argDesc[opt]; ok {
				c.Display = ui.T(c.Stem + " " + e + " (" + d + ")")
			} else {
				c.Display = ui.T(c.Stem + " (" + d + ")")
			}
		}
		return out.Put(c)
	}
	call := func(fn eval.Callable, args ...any) error {
		return fn.Call(fm, args, eval.NoOpts)
	}

	switch ctx.Type {
	case getopt.OptionOrArgument, getopt.Argument:
		// Find argument handler.
		var argHandler eval.Callable
		if len(parsedArgs) < len(argHandlers) {
			argHandler = argHandlers[len(parsedArgs)]
		} else if variadic {
			argHandler = argHandlers[len(argHandlers)-1]
		}
		if argHandler != nil {
			return call(argHandler, ctx.Text)
		}
		// TODO(xiaq): Notify that there is no suitable argument completer.
	case getopt.AnyOption:
		for _, opt := range opts.opts {
			if opt.Short != 0 {
				err := putShortOpt(opt)
				if err != nil {
					return err
				}
			}
			if opt.Long != "" {
				err := putLongOpt(opt)
				if err != nil {
					return err
				}
			}
		}
	case getopt.LongOption:
		for _, opt := range opts.opts {
			if opt.Long != "" && strings.HasPrefix(opt.Long, ctx.Text) {
				err := putLongOpt(opt)
				if err != nil {
					return err
				}
			}
		}
	case getopt.ChainShortOption:
		for _, opt := range opts.opts {
			if opt.Short != 0 {
				// TODO(xiaq): Loses chained options.
				err := putShortOpt(opt)
				if err != nil {
					return err
				}
			}
		}
	case getopt.OptionArgument:
		gen := opts.argGenerator[ctx.Option.Spec]
		if gen != nil {
			return call(gen, ctx.Option.Argument)
		}
	}
	return nil
}

// TODO(xiaq): Simplify most of the parsing below with reflection.

func parseGetoptArgs(v any) ([]string, error) {
	var args []string
	var err error
	errIterate := vals.Iterate(v, func(v any) bool {
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
	opts         []*getopt.OptionSpec
	desc         map[*getopt.OptionSpec]string
	argDesc      map[*getopt.OptionSpec]string
	argGenerator map[*getopt.OptionSpec]eval.Callable
}

func parseGetoptOptSpecs(v any) (parsedOptSpecs, error) {
	result := parsedOptSpecs{
		nil, map[*getopt.OptionSpec]string{},
		map[*getopt.OptionSpec]string{}, map[*getopt.OptionSpec]eval.Callable{}}

	var err error
	errIterate := vals.Iterate(v, func(v any) bool {
		m, ok := v.(vals.Map)
		if !ok {
			err = fmt.Errorf("opt should be map, got %s", vals.Kind(v))
			return false
		}

		opt := &getopt.OptionSpec{}

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
					"short should be exactly one rune, got %v", parse.Quote(s))
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
			opt.Arity = getopt.RequiredArgument
		case argOptional:
			opt.Arity = getopt.OptionalArgument
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

func parseGetoptArgHandlers(v any) ([]eval.Callable, bool, error) {
	var argHandlers []eval.Callable
	var variadic bool
	var err error
	errIterate := vals.Iterate(v, func(v any) bool {
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

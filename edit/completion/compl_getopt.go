package completion

import (
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/getopt"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hashmap"
)

func complGetopt(fm *eval.Frame, elemsv, optsv, argsv interface{}) {
	var (
		elems    []string
		opts     []*getopt.Option
		args     []eval.Callable
		variadic bool
	)
	desc := make(map[*getopt.Option]string)
	argdesc := make(map[*getopt.Option]string)
	// Convert arguments.
	err := vals.Iterate(elemsv, func(v interface{}) bool {
		elem, ok := v.(string)
		if !ok {
			throwf("arg should be string, got %s", vals.Kind(v))
		}
		elems = append(elems, elem)
		return true
	})
	maybeThrow(err)
	err = vals.Iterate(optsv, func(v interface{}) bool {
		m, ok := v.(hashmap.Map)
		if !ok {
			throwf("opt should be map, got %s", vals.Kind(v))
		}

		opt := &getopt.Option{}

		getStringField := func(k string) (string, bool) {
			v, ok := m.Index(k)
			if !ok {
				return "", false
			}
			if vs, ok := v.(string); ok {
				return vs, true
			}
			throwf("%s should be string, got %s", k, vals.Kind(v))
			panic("unreachable")
		}
		if s, ok := getStringField("short"); ok {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || size != len(s) {
				throwf("short option should be exactly one rune, got %v", parse.Quote(s))
			}
			opt.Short = r
		}
		if s, ok := getStringField("long"); ok {
			opt.Long = s
		}
		if opt.Short == 0 && opt.Long == "" {
			throwf("opt should have at least one of short and long forms")
		}

		getBoolField := func(k string) (bool, bool) {
			v, ok := m.Index(k)
			if !ok {
				return false, false
			}
			if vb, ok := v.(bool); ok {
				return vb, true
			}
			throwf("%s should be bool, got %s", k, vals.Kind(v))
			panic("unreachable")
		}
		argRequired, _ := getBoolField("arg-required")
		argOptional, _ := getBoolField("arg-optional")
		switch {
		case argRequired && argOptional:
			throwf("opt cannot have both arg-required and arg-optional")
		case argRequired:
			opt.HasArg = getopt.RequiredArgument
		case argOptional:
			opt.HasArg = getopt.OptionalArgument
		}

		if s, ok := getStringField("desc"); ok {
			desc[opt] = s
		}
		if s, ok := getStringField("arg-desc"); ok {
			argdesc[opt] = s
		}
		opts = append(opts, opt)
		return true
	})
	maybeThrow(err)
	err = vals.Iterate(argsv, func(v interface{}) bool {
		sv, ok := v.(string)
		if ok {
			if sv == "..." {
				variadic = true
				return true
			}
			throwf("string except for ... not allowed as argument handler, got %s", parse.Quote(sv))
		}
		arg, ok := v.(eval.Callable)
		if !ok {
			throwf("argument handler should be fn, got %s", vals.Kind(v))
		}
		args = append(args, arg)
		return true
	})
	maybeThrow(err)

	// TODO Configurable config
	g := getopt.Getopt{opts, getopt.GNUGetoptLong}
	_, parsedArgs, ctx := g.Parse(elems)
	out := fm.OutputChan()

	putShortOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "-" + string(opt.Short)}
		if d, ok := desc[opt]; ok {
			if e, ok := argdesc[opt]; ok {
				c.displaySuffix = " " + e + " (" + d + ")"
			} else {
				c.displaySuffix = " (" + d + ")"
			}
		}
		out <- c
	}
	putLongOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "--" + opt.Long}
		if d, ok := desc[opt]; ok {
			if e, ok := argdesc[opt]; ok {
				c.displaySuffix = " " + e + " (" + d + ")"
			} else {
				c.displaySuffix = " (" + d + ")"
			}
		}
		out <- c
	}

	switch ctx.Type {
	case getopt.NewOptionOrArgument, getopt.Argument:
		// Find argument completer
		var argCompl eval.Callable
		if len(parsedArgs) < len(args) {
			argCompl = args[len(parsedArgs)]
		} else if variadic {
			argCompl = args[len(args)-1]
		}
		if argCompl != nil {
			rawCands := make(chan rawCandidate)
			defer close(rawCands)
			go func() {
				for rc := range rawCands {
					out <- rc
				}
			}()
			err := callArgCompleter(argCompl, fm.Evaler, []string{ctx.Text}, rawCands)
			maybeThrow(err)
		}
		// TODO Notify that there is no suitable argument completer
	case getopt.NewOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				putShortOpt(opt)
			}
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.NewLongOption:
		for _, opt := range opts {
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.LongOption:
		for _, opt := range opts {
			if strings.HasPrefix(opt.Long, ctx.Text) {
				putLongOpt(opt)
			}
		}
	case getopt.ChainShortOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				// XXX loses chained options
				putShortOpt(opt)
			}
		}
	case getopt.OptionArgument:
	}
}

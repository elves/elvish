package edit

import (
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/getopt"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hashmap"
)

func complGetopt(ec *eval.Frame, a []types.Value, o map[string]types.Value) {
	var elemsv, optsv, argsv types.Value
	eval.ScanArgs(a, &elemsv, &optsv, &argsv)
	eval.TakeNoOpt(o)

	var (
		elems    []string
		opts     []*getopt.Option
		args     []eval.Fn
		variadic bool
	)
	desc := make(map[*getopt.Option]string)
	// Convert arguments.
	err := types.Iterate(elemsv, func(v types.Value) bool {
		elem, ok := v.(string)
		if !ok {
			throwf("arg should be string, got %s", types.Kind(v))
		}
		elems = append(elems, elem)
		return true
	})
	maybeThrow(err)
	err = types.Iterate(optsv, func(v types.Value) bool {
		m, ok := v.(hashmap.Map)
		if !ok {
			throwf("opt should be map, got %s", types.Kind(v))
		}
		get := func(k string) (string, bool) {
			v, ok := m.Get(k)
			if !ok {
				return "", false
			}
			if vs, ok := v.(string); ok {
				return vs, true
			}
			throwf("%s should be string, got %s", k, types.Kind(v))
			panic("unreachable")
		}

		opt := &getopt.Option{}
		if s, ok := get("short"); ok {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || size != len(s) {
				throwf("short option should be exactly one rune, got %v", parse.Quote(s))
			}
			opt.Short = r
		}
		if s, ok := get("long"); ok {
			opt.Long = s
		}
		if opt.Short == 0 && opt.Long == "" {
			throwf("opt should have at least one of short and long forms")
		}
		if s, ok := get("desc"); ok {
			desc[opt] = s
		}
		opts = append(opts, opt)
		return true
	})
	maybeThrow(err)
	err = types.Iterate(argsv, func(v types.Value) bool {
		sv, ok := v.(string)
		if ok {
			if sv == "..." {
				variadic = true
				return true
			}
			throwf("string except for ... not allowed as argument handler, got %s", parse.Quote(sv))
		}
		arg, ok := v.(eval.Fn)
		if !ok {
			throwf("argument handler should be fn, got %s", types.Kind(v))
		}
		args = append(args, arg)
		return true
	})
	maybeThrow(err)

	// TODO Configurable config
	g := getopt.Getopt{opts, getopt.GNUGetoptLong}
	_, parsedArgs, ctx := g.Parse(elems)
	out := ec.OutputChan()

	putShortOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "-" + string(opt.Short)}
		if d, ok := desc[opt]; ok {
			c.displaySuffix = " (" + d + ")"
		}
		out <- c
	}
	putLongOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "--" + opt.Long}
		if d, ok := desc[opt]; ok {
			c.displaySuffix = " (" + d + ")"
		}
		out <- c
	}

	switch ctx.Type {
	case getopt.NewOptionOrArgument, getopt.Argument:
		// Find argument completer
		var argCompl eval.Fn
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
			err := callArgCompleter(argCompl, ec.Evaler, []string{ctx.Text}, rawCands)
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
